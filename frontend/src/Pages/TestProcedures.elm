module Pages.TestProcedures exposing (Model, Msg, init, update, view)

import API
import Html exposing (Html)
import Html.Attributes
import Http
import Json.Encode as Encode
import Material.Button as Button
import Material.Card as Card
import Material.DataTable as DataTable
import Material.Dialog as Dialog
import Material.LayoutGrid as LayoutGrid
import Material.TextField as TextField
import Material.Typography as Typography
import Time
import Types exposing (PaginatedResponse, TestProcedure, TestProcedureInput, TestStep)



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
    }


type alias EditDialogState =
    { procedure : TestProcedure
    , name : String
    , description : String
    , stepsJson : String
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
                    ( { model | createDialog = Just { dialog | name = name } }
                    , Cmd.none
                    )

                Nothing ->
                    ( model, Cmd.none )

        SetCreateDescription description ->
            case model.createDialog of
                Just dialog ->
                    ( { model | createDialog = Just { dialog | description = description } }
                    , Cmd.none
                    )

                Nothing ->
                    ( model, Cmd.none )

        SetCreateStepsJson stepsJson ->
            case model.createDialog of
                Just dialog ->
                    ( { model | createDialog = Just { dialog | stepsJson = stepsJson } }
                    , Cmd.none
                    )

                Nothing ->
                    ( model, Cmd.none )

        SubmitCreate ->
            case model.createDialog of
                Just dialog ->
                    ( { model | loading = True }
                    , API.createTestProcedure
                        model.projectId
                        { name = dialog.name
                        , description = dialog.description
                        , steps = []
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
                    ( { model | editDialog = Just { dialog | name = name } }
                    , Cmd.none
                    )

                Nothing ->
                    ( model, Cmd.none )

        SetEditDescription description ->
            case model.editDialog of
                Just dialog ->
                    ( { model | editDialog = Just { dialog | description = description } }
                    , Cmd.none
                    )

                Nothing ->
                    ( model, Cmd.none )

        SetEditStepsJson stepsJson ->
            case model.editDialog of
                Just dialog ->
                    ( { model | editDialog = Just { dialog | stepsJson = stepsJson } }
                    , Cmd.none
                    )

                Nothing ->
                    ( model, Cmd.none )

        SubmitEdit ->
            case model.editDialog of
                Just dialog ->
                    ( { model | loading = True }
                    , API.updateTestProcedure
                        model.projectId
                        dialog.procedure.id
                        { name = dialog.name
                        , description = dialog.description
                        , steps = dialog.procedure.steps
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
        [ LayoutGrid.layoutGrid []
            [ LayoutGrid.cell
                [ LayoutGrid.span12 ]
                [ Html.div
                    [ Html.Attributes.style "display" "flex"
                    , Html.Attributes.style "justify-content" "space-between"
                    , Html.Attributes.style "align-items" "center"
                    , Html.Attributes.style "margin-bottom" "20px"
                    ]
                    [ Html.h1 [ Typography.headline3 ] [ Html.text "Test Procedures" ]
                    , Button.raised
                        (Button.config |> Button.setOnClick (Just OpenCreateDialog))
                        "Create Procedure"
                    ]
                ]
            , LayoutGrid.cell
                [ LayoutGrid.span12 ]
                [ case model.error of
                    Just err ->
                        Html.div
                            [ Html.Attributes.style "color" "red"
                            , Html.Attributes.style "margin-bottom" "20px"
                            ]
                            [ Html.text err ]

                    Nothing ->
                        Html.text ""
                ]
            , LayoutGrid.cell
                [ LayoutGrid.span12 ]
                [ if model.loading && List.isEmpty model.procedures then
                    Html.div [] [ Html.text "Loading..." ]

                  else
                    viewProceduresTable model.procedures
                ]
            , LayoutGrid.cell
                [ LayoutGrid.span12 ]
                [ viewPagination model ]
            ]
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
    DataTable.dataTable
        (DataTable.config |> DataTable.setAttributes [ Typography.typography ])
        { thead =
            [ DataTable.row []
                [ DataTable.cell [] [ Html.text "Name" ]
                , DataTable.cell [] [ Html.text "Description" ]
                , DataTable.cell [] [ Html.text "Version" ]
                , DataTable.cell [] [ Html.text "Steps" ]
                , DataTable.cell [] [ Html.text "Actions" ]
                ]
            ]
        , tbody =
            List.map viewProcedureRow procedures
        }


viewProcedureRow : TestProcedure -> DataTable.Row Msg
viewProcedureRow procedure =
    DataTable.row []
        [ DataTable.cell [] [ Html.text procedure.name ]
        , DataTable.cell [] [ Html.text procedure.description ]
        , DataTable.cell []
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
        , DataTable.cell [] [ Html.text (String.fromInt (List.length procedure.steps)) ]
        , DataTable.cell []
            [ Button.text
                (Button.config |> Button.setOnClick (Just (SelectProcedure procedure)))
                "View"
            , Button.text
                (Button.config |> Button.setOnClick (Just (OpenEditDialog procedure)))
                "Edit"
            , Button.text
                (Button.config |> Button.setOnClick (Just (CreateVersion procedure.id)))
                "New Version"
            , Button.text
                (Button.config |> Button.setOnClick (Just (OpenVersionsDialog procedure)))
                "History"
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
        [ Button.text
            (Button.config
                |> Button.setOnClick
                    (if hasPrev then
                        Just (LoadPage ((currentPage - 1) * model.limit))

                     else
                        Nothing
                    )
                |> Button.setDisabled (not hasPrev)
            )
            "Previous"
        , Html.span []
            [ Html.text
                ("Page "
                    ++ String.fromInt (currentPage + 1)
                    ++ " of "
                    ++ String.fromInt (max 1 totalPages)
                )
            ]
        , Button.text
            (Button.config
                |> Button.setOnClick
                    (if hasNext then
                        Just (LoadPage ((currentPage + 1) * model.limit))

                     else
                        Nothing
                    )
                |> Button.setDisabled (not hasNext)
            )
            "Next"
        ]


viewCreateDialog : CreateDialogState -> Html Msg
viewCreateDialog dialog =
    Dialog.dialog
        (Dialog.config
            |> Dialog.setOpen True
            |> Dialog.setOnClose CloseCreateDialog
        )
        { title = Just "Create Test Procedure"
        , content =
            [ Html.div []
                [ TextField.filled
                    (TextField.config
                        |> TextField.setLabel (Just "Name")
                        |> TextField.setValue (Just dialog.name)
                        |> TextField.setOnInput (Just SetCreateName)
                        |> TextField.setRequired True
                    )
                , TextField.filled
                    (TextField.config
                        |> TextField.setLabel (Just "Description")
                        |> TextField.setValue (Just dialog.description)
                        |> TextField.setOnInput (Just SetCreateDescription)
                        |> TextField.setRequired True
                    )
                , TextField.filled
                    (TextField.config
                        |> TextField.setLabel (Just "Steps (JSON)")
                        |> TextField.setValue (Just dialog.stepsJson)
                        |> TextField.setOnInput (Just SetCreateStepsJson)
                        |> TextField.setRequired True
                    )
                ]
            ]
        , actions =
            [ Button.text
                (Button.config |> Button.setOnClick (Just CloseCreateDialog))
                "Cancel"
            , Button.raised
                (Button.config |> Button.setOnClick (Just SubmitCreate))
                "Create"
            ]
        }


viewEditDialog : EditDialogState -> Html Msg
viewEditDialog dialog =
    Dialog.dialog
        (Dialog.config
            |> Dialog.setOpen True
            |> Dialog.setOnClose CloseEditDialog
        )
        { title = Just "Edit Test Procedure"
        , content =
            [ Html.div []
                [ TextField.filled
                    (TextField.config
                        |> TextField.setLabel (Just "Name")
                        |> TextField.setValue (Just dialog.name)
                        |> TextField.setOnInput (Just SetEditName)
                        |> TextField.setRequired True
                    )
                , TextField.filled
                    (TextField.config
                        |> TextField.setLabel (Just "Description")
                        |> TextField.setValue (Just dialog.description)
                        |> TextField.setOnInput (Just SetEditDescription)
                        |> TextField.setRequired True
                    )
                , TextField.filled
                    (TextField.config
                        |> TextField.setLabel (Just "Steps (JSON)")
                        |> TextField.setValue (Just dialog.stepsJson)
                        |> TextField.setOnInput (Just SetEditStepsJson)
                        |> TextField.setRequired True
                    )
                ]
            ]
        , actions =
            [ Button.text
                (Button.config |> Button.setOnClick (Just CloseEditDialog))
                "Cancel"
            , Button.raised
                (Button.config |> Button.setOnClick (Just SubmitEdit))
                "Save"
            ]
        }


viewVersionsDialog : VersionsDialogState -> Html Msg
viewVersionsDialog dialog =
    Dialog.dialog
        (Dialog.config
            |> Dialog.setOpen True
            |> Dialog.setOnClose CloseVersionsDialog
        )
        { title = Just ("Version History: " ++ dialog.procedure.name)
        , content =
            [ Html.div []
                [ if List.isEmpty dialog.versions then
                    Html.text "Loading versions..."

                  else
                    DataTable.dataTable
                        (DataTable.config)
                        { thead =
                            [ DataTable.row []
                                [ DataTable.cell [] [ Html.text "Version" ]
                                , DataTable.cell [] [ Html.text "Created" ]
                                , DataTable.cell [] [ Html.text "Steps" ]
                                ]
                            ]
                        , tbody =
                            List.map
                                (\v ->
                                    DataTable.row []
                                        [ DataTable.cell []
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
                                        , DataTable.cell [] [ Html.text (formatTime v.createdAt) ]
                                        , DataTable.cell [] [ Html.text (String.fromInt (List.length v.steps)) ]
                                        ]
                                )
                                dialog.versions
                        }
                ]
            ]
        , actions =
            [ Button.text
                (Button.config |> Button.setOnClick (Just CloseVersionsDialog))
                "Close"
            ]
        }



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


stepsToJson : List TestStep -> String
stepsToJson steps =
    "[]"
