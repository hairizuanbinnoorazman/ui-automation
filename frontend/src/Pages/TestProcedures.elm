module Pages.TestProcedures exposing (Model, Msg, init, update, view)

import API
import Components
import Html exposing (Html)
import Html.Attributes
import Html.Events
import Http
import Json.Decode as Decode
import Json.Encode as Encode
import Time
import Types exposing (PaginatedResponse, TestProcedure, TestProcedureInput, TestStep, testStepDecoder, testStepEncoder)



-- MODEL


type alias Model =
    { projectId : String
    , procedures : List TestProcedure
    , total : Int
    , limit : Int
    , offset : Int
    , loading : Bool
    , error : Maybe String
    , createDialog : Maybe CreateDialogState
    , editDialog : Maybe EditDialogState
    , versionsDialog : Maybe VersionsDialogState
    , selectedProcedure : Maybe TestProcedure
    }


type alias CreateDialogState =
    { name : String
    , description : String
    , stepsJson : String
    , error : Maybe String
    }


type alias EditDialogState =
    { procedure : TestProcedure
    , name : String
    , description : String
    , stepsJson : String
    , error : Maybe String
    }


type alias VersionsDialogState =
    { procedure : TestProcedure
    , versions : List TestProcedure
    }


init : String -> ( Model, Cmd Msg )
init projectId =
    ( { projectId = projectId
      , procedures = []
      , total = 0
      , limit = 10
      , offset = 0
      , loading = False
      , error = Nothing
      , createDialog = Nothing
      , editDialog = Nothing
      , versionsDialog = Nothing
      , selectedProcedure = Nothing
      }
    , API.getTestProcedures projectId 10 0 ProceduresResponse
    )



-- UPDATE


type Msg
    = ProceduresResponse (Result Http.Error (PaginatedResponse TestProcedure))
    | LoadPage Int
    | SelectProcedure TestProcedure
    | OpenCreateDialog
    | CloseCreateDialog
    | SetCreateName String
    | SetCreateDescription String
    | SetCreateStepsJson String
    | SubmitCreate
    | CreateResponse (Result Http.Error TestProcedure)
    | OpenEditDialog TestProcedure
    | CloseEditDialog
    | SetEditName String
    | SetEditDescription String
    | SetEditStepsJson String
    | SubmitEdit
    | EditResponse (Result Http.Error TestProcedure)
    | CreateVersion String
    | CreateVersionResponse (Result Http.Error TestProcedure)
    | OpenVersionsDialog TestProcedure
    | CloseVersionsDialog
    | VersionsResponse (Result Http.Error (List TestProcedure))


update : Msg -> Model -> ( Model, Cmd Msg )
update msg model =
    case msg of
        ProceduresResponse (Ok response) ->
            ( { model
                | procedures = response.items
                , total = response.total
                , loading = False
                , error = Nothing
              }
            , Cmd.none
            )

        ProceduresResponse (Err error) ->
            ( { model
                | loading = False
                , error = Just (httpErrorToString error)
              }
            , Cmd.none
            )

        LoadPage offset ->
            ( { model | loading = True, offset = offset }
            , API.getTestProcedures model.projectId model.limit offset ProceduresResponse
            )

        SelectProcedure procedure ->
            ( { model | selectedProcedure = Just procedure }
            , Cmd.none
            )

        OpenCreateDialog ->
            ( { model
                | createDialog =
                    Just
                        { name = ""
                        , description = ""
                        , stepsJson = "[]"
                        , error = Nothing
                        }
              }
            , Cmd.none
            )

        CloseCreateDialog ->
            ( { model | createDialog = Nothing }
            , Cmd.none
            )

        SetCreateName name ->
            case model.createDialog of
                Just dialog ->
                    ( { model | createDialog = Just { dialog | name = name, error = Nothing } }
                    , Cmd.none
                    )

                Nothing ->
                    ( model, Cmd.none )

        SetCreateDescription description ->
            case model.createDialog of
                Just dialog ->
                    ( { model | createDialog = Just { dialog | description = description, error = Nothing } }
                    , Cmd.none
                    )

                Nothing ->
                    ( model, Cmd.none )

        SetCreateStepsJson stepsJson ->
            case model.createDialog of
                Just dialog ->
                    ( { model | createDialog = Just { dialog | stepsJson = stepsJson, error = Nothing } }
                    , Cmd.none
                    )

                Nothing ->
                    ( model, Cmd.none )

        SubmitCreate ->
            case model.createDialog of
                Just dialog ->
                    case parseStepsJson dialog.stepsJson of
                        Err err ->
                            ( { model
                                | createDialog = Just { dialog | error = Just ("Invalid steps JSON: " ++ Decode.errorToString err) }
                              }
                            , Cmd.none
                            )

                        Ok parsedSteps ->
                            ( { model | loading = True }
                            , API.createTestProcedure
                                model.projectId
                                { name = dialog.name
                                , description = dialog.description
                                , steps = parsedSteps
                                }
                                CreateResponse
                            )

                Nothing ->
                    ( model, Cmd.none )

        CreateResponse (Ok procedure) ->
            ( { model
                | loading = False
                , createDialog = Nothing
              }
            , API.getTestProcedures model.projectId model.limit model.offset ProceduresResponse
            )

        CreateResponse (Err error) ->
            ( { model
                | loading = False
                , error = Just (httpErrorToString error)
              }
            , Cmd.none
            )

        OpenEditDialog procedure ->
            ( { model
                | editDialog =
                    Just
                        { procedure = procedure
                        , name = procedure.name
                        , description = procedure.description
                        , stepsJson = stepsToJson procedure.steps
                        , error = Nothing
                        }
              }
            , Cmd.none
            )

        CloseEditDialog ->
            ( { model | editDialog = Nothing }
            , Cmd.none
            )

        SetEditName name ->
            case model.editDialog of
                Just dialog ->
                    ( { model | editDialog = Just { dialog | name = name, error = Nothing } }
                    , Cmd.none
                    )

                Nothing ->
                    ( model, Cmd.none )

        SetEditDescription description ->
            case model.editDialog of
                Just dialog ->
                    ( { model | editDialog = Just { dialog | description = description, error = Nothing } }
                    , Cmd.none
                    )

                Nothing ->
                    ( model, Cmd.none )

        SetEditStepsJson stepsJson ->
            case model.editDialog of
                Just dialog ->
                    ( { model | editDialog = Just { dialog | stepsJson = stepsJson, error = Nothing } }
                    , Cmd.none
                    )

                Nothing ->
                    ( model, Cmd.none )

        SubmitEdit ->
            case model.editDialog of
                Just dialog ->
                    case parseStepsJson dialog.stepsJson of
                        Err err ->
                            ( { model
                                | editDialog = Just { dialog | error = Just ("Invalid steps JSON: " ++ Decode.errorToString err) }
                              }
                            , Cmd.none
                            )

                        Ok parsedSteps ->
                            ( { model | loading = True }
                            , API.updateTestProcedure
                                model.projectId
                                dialog.procedure.id
                                { name = dialog.name
                                , description = dialog.description
                                , steps = parsedSteps
                                }
                                EditResponse
                            )

                Nothing ->
                    ( model, Cmd.none )

        EditResponse (Ok procedure) ->
            ( { model
                | loading = False
                , editDialog = Nothing
              }
            , API.getTestProcedures model.projectId model.limit model.offset ProceduresResponse
            )

        EditResponse (Err error) ->
            ( { model
                | loading = False
                , error = Just (httpErrorToString error)
              }
            , Cmd.none
            )

        CreateVersion procedureId ->
            ( { model | loading = True }
            , API.createProcedureVersion model.projectId procedureId CreateVersionResponse
            )

        CreateVersionResponse (Ok procedure) ->
            ( { model | loading = False }
            , API.getTestProcedures model.projectId model.limit model.offset ProceduresResponse
            )

        CreateVersionResponse (Err error) ->
            ( { model
                | loading = False
                , error = Just (httpErrorToString error)
              }
            , Cmd.none
            )

        OpenVersionsDialog procedure ->
            ( { model
                | versionsDialog =
                    Just
                        { procedure = procedure
                        , versions = []
                        }
              }
            , API.getProcedureVersions model.projectId procedure.id VersionsResponse
            )

        CloseVersionsDialog ->
            ( { model | versionsDialog = Nothing }
            , Cmd.none
            )

        VersionsResponse (Ok versions) ->
            case model.versionsDialog of
                Just dialog ->
                    ( { model
                        | versionsDialog = Just { dialog | versions = versions }
                      }
                    , Cmd.none
                    )

                Nothing ->
                    ( model, Cmd.none )

        VersionsResponse (Err error) ->
            ( { model
                | error = Just (httpErrorToString error)
              }
            , Cmd.none
            )



-- VIEW


view : Model -> Html Msg
view model =
    Html.div []
        [ Html.div
            [ Html.Attributes.class "page-header"
            ]
            [ Html.h1 [ Html.Attributes.class "mdc-typography--headline3" ] [ Html.text "Test Procedures" ]
            , Html.button
                [ Html.Events.onClick OpenCreateDialog
                , Html.Attributes.class "mdc-button mdc-button--raised"
                ]
                [ Html.text "Create Procedure" ]
            ]
        , case model.error of
            Just err ->
                Html.div
                    [ Html.Attributes.style "color" "red"
                    , Html.Attributes.style "margin-bottom" "20px"
                    ]
                    [ Html.text err ]

            Nothing ->
                Html.text ""
        , if model.loading && List.isEmpty model.procedures then
            Html.div [] [ Html.text "Loading..." ]

          else
            viewProceduresTable model.procedures
        , viewPagination model
        , case model.createDialog of
            Just dialog ->
                viewCreateDialog dialog

            Nothing ->
                Html.text ""
        , case model.editDialog of
            Just dialog ->
                viewEditDialog dialog

            Nothing ->
                Html.text ""
        , case model.versionsDialog of
            Just dialog ->
                viewVersionsDialog dialog

            Nothing ->
                Html.text ""
        ]


viewProceduresTable : List TestProcedure -> Html Msg
viewProceduresTable procedures =
    Html.table
        [ Html.Attributes.class "mdc-data-table__table"
        , Html.Attributes.style "width" "100%"
        , Html.Attributes.style "border-collapse" "collapse"
        ]
        [ Html.thead []
            [ Html.tr []
                [ Html.th [ Html.Attributes.style "text-align" "left", Html.Attributes.style "padding" "12px" ] [ Html.text "Name" ]
                , Html.th [ Html.Attributes.style "text-align" "left", Html.Attributes.style "padding" "12px" ] [ Html.text "Description" ]
                , Html.th [ Html.Attributes.style "text-align" "left", Html.Attributes.style "padding" "12px" ] [ Html.text "Version" ]
                , Html.th [ Html.Attributes.style "text-align" "left", Html.Attributes.style "padding" "12px" ] [ Html.text "Steps" ]
                , Html.th [ Html.Attributes.style "text-align" "left", Html.Attributes.style "padding" "12px" ] [ Html.text "Actions" ]
                ]
            ]
        , Html.tbody []
            (List.map viewProcedureRow procedures)
        ]


viewProcedureRow : TestProcedure -> Html Msg
viewProcedureRow procedure =
    Html.tr [ Html.Attributes.style "border-bottom" "1px solid #ddd" ]
        [ Html.td [ Html.Attributes.style "padding" "12px" ] [ Html.text procedure.name ]
        , Html.td [ Html.Attributes.style "padding" "12px" ] [ Html.text procedure.description ]
        , Html.td [ Html.Attributes.style "padding" "12px" ]
            [ Html.text
                ("v"
                    ++ String.fromInt procedure.version
                    ++ (if procedure.isLatest then
                            " (latest)"

                        else
                            ""
                       )
                )
            ]
        , Html.td [ Html.Attributes.style "padding" "12px" ] [ Html.text (String.fromInt (List.length procedure.steps)) ]
        , Html.td [ Html.Attributes.style "padding" "12px" ]
            [ Html.button
                [ Html.Events.onClick (SelectProcedure procedure)
                , Html.Attributes.class "mdc-button"
                , Html.Attributes.style "margin-right" "8px"
                ]
                [ Html.text "View" ]
            , Html.button
                [ Html.Events.onClick (OpenEditDialog procedure)
                , Html.Attributes.class "mdc-button"
                , Html.Attributes.style "margin-right" "8px"
                ]
                [ Html.text "Edit" ]
            , Html.button
                [ Html.Events.onClick (CreateVersion procedure.id)
                , Html.Attributes.class "mdc-button"
                , Html.Attributes.style "margin-right" "8px"
                ]
                [ Html.text "New Version" ]
            , Html.button
                [ Html.Events.onClick (OpenVersionsDialog procedure)
                , Html.Attributes.class "mdc-button"
                ]
                [ Html.text "History" ]
            ]
        ]


viewPagination : Model -> Html Msg
viewPagination model =
    let
        currentPage =
            model.offset // model.limit

        totalPages =
            (model.total + model.limit - 1) // model.limit

        hasPrev =
            currentPage > 0

        hasNext =
            currentPage < totalPages - 1
    in
    Html.div
        [ Html.Attributes.style "display" "flex"
        , Html.Attributes.style "justify-content" "center"
        , Html.Attributes.style "align-items" "center"
        , Html.Attributes.style "gap" "10px"
        , Html.Attributes.style "margin-top" "20px"
        ]
        [ Html.button
            [ Html.Events.onClick (LoadPage ((currentPage - 1) * model.limit))
            , Html.Attributes.disabled (not hasPrev)
            , Html.Attributes.class "mdc-button"
            ]
            [ Html.text "Previous" ]
        , Html.span []
            [ Html.text
                ("Page "
                    ++ String.fromInt (currentPage + 1)
                    ++ " of "
                    ++ String.fromInt (max 1 totalPages)
                )
            ]
        , Html.button
            [ Html.Events.onClick (LoadPage ((currentPage + 1) * model.limit))
            , Html.Attributes.disabled (not hasNext)
            , Html.Attributes.class "mdc-button"
            ]
            [ Html.text "Next" ]
        ]


viewCreateDialog : CreateDialogState -> Html Msg
viewCreateDialog dialog =
    Components.viewDialogOverlay "Create Test Procedure"
        [ case dialog.error of
            Just err ->
                Html.div
                    [ Html.Attributes.style "color" "red"
                    , Html.Attributes.style "background-color" "#ffebee"
                    , Html.Attributes.style "padding" "12px"
                    , Html.Attributes.style "border-radius" "4px"
                    , Html.Attributes.style "margin-bottom" "16px"
                    , Html.Attributes.style "border" "1px solid #ffcdd2"
                    ]
                    [ Html.text err ]

            Nothing ->
                Html.text ""
        , Components.viewFormField "Name"
            [ Html.Attributes.type_ "text"
            , Html.Attributes.value dialog.name
            , Html.Events.onInput SetCreateName
            , Html.Attributes.required True
            ]
        , Components.viewFormField "Description"
            [ Html.Attributes.type_ "text"
            , Html.Attributes.value dialog.description
            , Html.Events.onInput SetCreateDescription
            , Html.Attributes.required True
            ]
        , Components.viewFormField "Steps (JSON)"
            [ Html.Attributes.type_ "text"
            , Html.Attributes.value dialog.stepsJson
            , Html.Events.onInput SetCreateStepsJson
            , Html.Attributes.required True
            ]
        ]
        [ Html.button
            [ Html.Events.onClick CloseCreateDialog
            , Html.Attributes.class "mdc-button"
            ]
            [ Html.text "Cancel" ]
        , Html.button
            [ Html.Events.onClick SubmitCreate
            , Html.Attributes.class "mdc-button mdc-button--raised"
            ]
            [ Html.text "Create" ]
        ]


viewEditDialog : EditDialogState -> Html Msg
viewEditDialog dialog =
    Components.viewDialogOverlay "Edit Test Procedure"
        [ case dialog.error of
            Just err ->
                Html.div
                    [ Html.Attributes.style "color" "red"
                    , Html.Attributes.style "background-color" "#ffebee"
                    , Html.Attributes.style "padding" "12px"
                    , Html.Attributes.style "border-radius" "4px"
                    , Html.Attributes.style "margin-bottom" "16px"
                    , Html.Attributes.style "border" "1px solid #ffcdd2"
                    ]
                    [ Html.text err ]

            Nothing ->
                Html.text ""
        , Components.viewFormField "Name"
            [ Html.Attributes.type_ "text"
            , Html.Attributes.value dialog.name
            , Html.Events.onInput SetEditName
            , Html.Attributes.required True
            ]
        , Components.viewFormField "Description"
            [ Html.Attributes.type_ "text"
            , Html.Attributes.value dialog.description
            , Html.Events.onInput SetEditDescription
            , Html.Attributes.required True
            ]
        , Components.viewFormField "Steps (JSON)"
            [ Html.Attributes.type_ "text"
            , Html.Attributes.value dialog.stepsJson
            , Html.Events.onInput SetEditStepsJson
            , Html.Attributes.required True
            ]
        ]
        [ Html.button
            [ Html.Events.onClick CloseEditDialog
            , Html.Attributes.class "mdc-button"
            ]
            [ Html.text "Cancel" ]
        , Html.button
            [ Html.Events.onClick SubmitEdit
            , Html.Attributes.class "mdc-button mdc-button--raised"
            ]
            [ Html.text "Save" ]
        ]


viewVersionsDialog : VersionsDialogState -> Html Msg
viewVersionsDialog dialog =
    Components.viewDialogOverlay ("Version History: " ++ dialog.procedure.name)
        [ Html.div []
            [ if List.isEmpty dialog.versions then
                Html.text "Loading versions..."

              else
                Html.table
                    [ Html.Attributes.class "mdc-data-table__table"
                    , Html.Attributes.style "width" "100%"
                    , Html.Attributes.style "border-collapse" "collapse"
                    ]
                    [ Html.thead []
                        [ Html.tr []
                            [ Html.th [ Html.Attributes.style "text-align" "left", Html.Attributes.style "padding" "12px" ] [ Html.text "Version" ]
                            , Html.th [ Html.Attributes.style "text-align" "left", Html.Attributes.style "padding" "12px" ] [ Html.text "Created" ]
                            , Html.th [ Html.Attributes.style "text-align" "left", Html.Attributes.style "padding" "12px" ] [ Html.text "Steps" ]
                            ]
                        ]
                    , Html.tbody []
                        (List.map
                            (\v ->
                                Html.tr [ Html.Attributes.style "border-bottom" "1px solid #ddd" ]
                                    [ Html.td [ Html.Attributes.style "padding" "12px" ]
                                        [ Html.text
                                            ("v"
                                                ++ String.fromInt v.version
                                                ++ (if v.isLatest then
                                                        " (latest)"

                                                    else
                                                        ""
                                                   )
                                            )
                                        ]
                                    , Html.td [ Html.Attributes.style "padding" "12px" ] [ Html.text (formatTime v.createdAt) ]
                                    , Html.td [ Html.Attributes.style "padding" "12px" ] [ Html.text (String.fromInt (List.length v.steps)) ]
                                    ]
                            )
                            dialog.versions
                        )
                    ]
            ]
        ]
        [ Html.button
            [ Html.Events.onClick CloseVersionsDialog
            , Html.Attributes.class "mdc-button"
            ]
            [ Html.text "Close" ]
        ]



-- HELPERS


httpErrorToString : Http.Error -> String
httpErrorToString error =
    case error of
        Http.BadUrl _ ->
            "Invalid URL"

        Http.Timeout ->
            "Request timed out"

        Http.NetworkError ->
            "Network error"

        Http.BadStatus status ->
            "Server error: " ++ String.fromInt status

        Http.BadBody body ->
            "Invalid response: " ++ body


formatTime : Time.Posix -> String
formatTime time =
    let
        year =
            String.fromInt (Time.toYear Time.utc time)

        month =
            String.fromInt (monthToInt (Time.toMonth Time.utc time))

        day =
            String.fromInt (Time.toDay Time.utc time)
    in
    year ++ "-" ++ String.padLeft 2 '0' month ++ "-" ++ String.padLeft 2 '0' day


monthToInt : Time.Month -> Int
monthToInt month =
    case month of
        Time.Jan ->
            1

        Time.Feb ->
            2

        Time.Mar ->
            3

        Time.Apr ->
            4

        Time.May ->
            5

        Time.Jun ->
            6

        Time.Jul ->
            7

        Time.Aug ->
            8

        Time.Sep ->
            9

        Time.Oct ->
            10

        Time.Nov ->
            11

        Time.Dec ->
            12


-- Decode array of steps from JSON string
stepsDecoder : Decode.Decoder (List TestStep)
stepsDecoder =
    Decode.list testStepDecoder


-- Parse JSON string to List TestStep
parseStepsJson : String -> Result Decode.Error (List TestStep)
parseStepsJson jsonString =
    let
        trimmed =
            String.trim jsonString
    in
    if String.isEmpty trimmed then
        Ok []

    else
        Decode.decodeString stepsDecoder trimmed


-- Convert list of steps to formatted JSON string
stepsToJson : List TestStep -> String
stepsToJson steps =
    steps
        |> Encode.list testStepEncoder
        |> Encode.encode 2
