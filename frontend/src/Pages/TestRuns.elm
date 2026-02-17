module Pages.TestRuns exposing (Model, Msg, init, update, view)

import API
import Components
import Html exposing (Html)
import Html.Attributes
import Html.Events
import Http
import Time
import Types exposing (AssetType, CompleteTestRunInput, PaginatedResponse, TestRun, TestRunAsset, TestRunInput, TestRunStatus)



-- MODEL


type alias Model =
    { procedureId : String
    , runs : List TestRun
    , selectedRun : Maybe TestRun
    , assets : List TestRunAsset
    , total : Int
    , limit : Int
    , offset : Int
    , loading : Bool
    , error : Maybe String
    , createDialog : Maybe CreateDialogState
    , completeDialog : Maybe CompleteDialogState
    }


type alias CreateDialogState =
    { notes : String
    }


type alias CompleteDialogState =
    { run : TestRun
    , status : TestRunStatus
    , notes : String
    }


init : String -> ( Model, Cmd Msg )
init procedureId =
    ( { procedureId = procedureId
      , runs = []
      , selectedRun = Nothing
      , assets = []
      , total = 0
      , limit = 10
      , offset = 0
      , loading = False
      , error = Nothing
      , createDialog = Nothing
      , completeDialog = Nothing
      }
    , API.getTestRuns procedureId 10 0 RunsResponse
    )



-- UPDATE


type Msg
    = RunsResponse (Result Http.Error (PaginatedResponse TestRun))
    | LoadPage Int
    | SelectRun TestRun
    | AssetsResponse (Result Http.Error (List TestRunAsset))
    | OpenCreateDialog
    | CloseCreateDialog
    | SetCreateNotes String
    | SubmitCreate
    | CreateResponse (Result Http.Error TestRun)
    | StartRun String
    | StartRunResponse (Result Http.Error TestRun)
    | OpenCompleteDialog TestRun
    | CloseCompleteDialog
    | SetCompleteStatus String
    | SetCompleteNotes String
    | SubmitComplete
    | CompleteResponse (Result Http.Error TestRun)
    | DeleteAsset String String
    | DeleteAssetResponse (Result Http.Error ())


update : Msg -> Model -> ( Model, Cmd Msg )
update msg model =
    case msg of
        RunsResponse (Ok response) ->
            ( { model
                | runs = response.items
                , total = response.total
                , loading = False
                , error = Nothing
              }
            , Cmd.none
            )

        RunsResponse (Err error) ->
            ( { model
                | loading = False
                , error = Just (httpErrorToString error)
              }
            , Cmd.none
            )

        LoadPage offset ->
            ( { model | loading = True, offset = offset }
            , API.getTestRuns model.procedureId model.limit offset RunsResponse
            )

        SelectRun run ->
            ( { model | selectedRun = Just run, assets = [] }
            , API.getTestRunAssets run.id AssetsResponse
            )

        AssetsResponse (Ok assets) ->
            ( { model | assets = assets }
            , Cmd.none
            )

        AssetsResponse (Err error) ->
            ( { model
                | error = Just (httpErrorToString error)
              }
            , Cmd.none
            )

        OpenCreateDialog ->
            ( { model
                | createDialog = Just { notes = "" }
              }
            , Cmd.none
            )

        CloseCreateDialog ->
            ( { model | createDialog = Nothing }
            , Cmd.none
            )

        SetCreateNotes notes ->
            case model.createDialog of
                Just dialog ->
                    ( { model | createDialog = Just { dialog | notes = notes } }
                    , Cmd.none
                    )

                Nothing ->
                    ( model, Cmd.none )

        SubmitCreate ->
            case model.createDialog of
                Just dialog ->
                    ( { model | loading = True }
                    , API.createTestRun
                        model.procedureId
                        { notes = dialog.notes }
                        CreateResponse
                    )

                Nothing ->
                    ( model, Cmd.none )

        CreateResponse (Ok run) ->
            ( { model
                | loading = False
                , createDialog = Nothing
              }
            , API.getTestRuns model.procedureId model.limit model.offset RunsResponse
            )

        CreateResponse (Err error) ->
            ( { model
                | loading = False
                , error = Just (httpErrorToString error)
              }
            , Cmd.none
            )

        StartRun runId ->
            ( { model | loading = True }
            , API.startTestRun runId StartRunResponse
            )

        StartRunResponse (Ok run) ->
            ( { model | loading = False }
            , API.getTestRuns model.procedureId model.limit model.offset RunsResponse
            )

        StartRunResponse (Err error) ->
            ( { model
                | loading = False
                , error = Just (httpErrorToString error)
              }
            , Cmd.none
            )

        OpenCompleteDialog run ->
            ( { model
                | completeDialog =
                    Just
                        { run = run
                        , status = Types.Passed
                        , notes = run.notes
                        }
              }
            , Cmd.none
            )

        CloseCompleteDialog ->
            ( { model | completeDialog = Nothing }
            , Cmd.none
            )

        SetCompleteStatus statusStr ->
            case model.completeDialog of
                Just dialog ->
                    let
                        status =
                            stringToStatus statusStr
                    in
                    ( { model | completeDialog = Just { dialog | status = status } }
                    , Cmd.none
                    )

                Nothing ->
                    ( model, Cmd.none )

        SetCompleteNotes notes ->
            case model.completeDialog of
                Just dialog ->
                    ( { model | completeDialog = Just { dialog | notes = notes } }
                    , Cmd.none
                    )

                Nothing ->
                    ( model, Cmd.none )

        SubmitComplete ->
            case model.completeDialog of
                Just dialog ->
                    ( { model | loading = True }
                    , API.completeTestRun
                        dialog.run.id
                        { status = dialog.status
                        , notes = dialog.notes
                        }
                        CompleteResponse
                    )

                Nothing ->
                    ( model, Cmd.none )

        CompleteResponse (Ok run) ->
            ( { model
                | loading = False
                , completeDialog = Nothing
              }
            , API.getTestRuns model.procedureId model.limit model.offset RunsResponse
            )

        CompleteResponse (Err error) ->
            ( { model
                | loading = False
                , error = Just (httpErrorToString error)
              }
            , Cmd.none
            )

        DeleteAsset runId assetId ->
            ( { model | loading = True }
            , API.deleteTestRunAsset runId assetId DeleteAssetResponse
            )

        DeleteAssetResponse (Ok ()) ->
            case model.selectedRun of
                Just run ->
                    ( { model | loading = False }
                    , API.getTestRunAssets run.id AssetsResponse
                    )

                Nothing ->
                    ( { model | loading = False }
                    , Cmd.none
                    )

        DeleteAssetResponse (Err error) ->
            ( { model
                | loading = False
                , error = Just (httpErrorToString error)
              }
            , Cmd.none
            )



-- VIEW


view : Model -> Html Msg
view model =
    Html.div []
        [ Html.div
            [ Html.Attributes.style "display" "flex"
            , Html.Attributes.style "gap" "20px"
            ]
            [ Html.div
                [ Html.Attributes.style "flex" "2" ]
                [ Html.div
                    [ Html.Attributes.class "page-header"
                    ]
                    [ Html.h1 [ Html.Attributes.class "mdc-typography--headline3" ] [ Html.text "Test Runs" ]
                    , Html.button
                        [ Html.Events.onClick OpenCreateDialog
                        , Html.Attributes.class "mdc-button mdc-button--raised"
                        ]
                        [ Html.text "Create Run" ]
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
                , if model.loading && List.isEmpty model.runs then
                    Html.div [] [ Html.text "Loading..." ]

                  else
                    viewRunsTable model.runs
                , viewPagination model
                ]
            , Html.div
                [ Html.Attributes.style "flex" "1" ]
                [ case model.selectedRun of
                    Just run ->
                        viewRunDetails run model.assets

                    Nothing ->
                        Html.div []
                            [ Html.h2 [ Html.Attributes.class "mdc-typography--headline5" ] [ Html.text "Select a run to view details" ]
                            ]
                ]
            ]
        , case model.createDialog of
            Just dialog ->
                viewCreateDialog dialog

            Nothing ->
                Html.text ""
        , case model.completeDialog of
            Just dialog ->
                viewCompleteDialog dialog

            Nothing ->
                Html.text ""
        ]


viewRunsTable : List TestRun -> Html Msg
viewRunsTable runs =
    Html.table
        [ Html.Attributes.class "mdc-data-table__table"
        , Html.Attributes.style "width" "100%"
        , Html.Attributes.style "border-collapse" "collapse"
        ]
        [ Html.thead []
            [ Html.tr []
                [ Html.th [ Html.Attributes.style "text-align" "left", Html.Attributes.style "padding" "12px" ] [ Html.text "Status" ]
                , Html.th [ Html.Attributes.style "text-align" "left", Html.Attributes.style "padding" "12px" ] [ Html.text "Notes" ]
                , Html.th [ Html.Attributes.style "text-align" "left", Html.Attributes.style "padding" "12px" ] [ Html.text "Created" ]
                , Html.th [ Html.Attributes.style "text-align" "left", Html.Attributes.style "padding" "12px" ] [ Html.text "Actions" ]
                ]
            ]
        , Html.tbody []
            (List.map viewRunRow runs)
        ]


viewRunRow : TestRun -> Html Msg
viewRunRow run =
    Html.tr [ Html.Attributes.style "border-bottom" "1px solid #ddd" ]
        [ Html.td [ Html.Attributes.style "padding" "12px" ] [ Html.text (statusToString run.status) ]
        , Html.td [ Html.Attributes.style "padding" "12px" ] [ Html.text run.notes ]
        , Html.td [ Html.Attributes.style "padding" "12px" ] [ Html.text (formatTime run.createdAt) ]
        , Html.td [ Html.Attributes.style "padding" "12px" ]
            [ Html.button
                [ Html.Events.onClick (SelectRun run)
                , Html.Attributes.class "mdc-button"
                , Html.Attributes.style "margin-right" "8px"
                ]
                [ Html.text "View" ]
            , if run.status == Types.Pending then
                Html.button
                    [ Html.Events.onClick (StartRun run.id)
                    , Html.Attributes.class "mdc-button"
                    ]
                    [ Html.text "Start" ]

              else if run.status == Types.Running then
                Html.button
                    [ Html.Events.onClick (OpenCompleteDialog run)
                    , Html.Attributes.class "mdc-button"
                    ]
                    [ Html.text "Complete" ]

              else
                Html.text ""
            ]
        ]


viewRunDetails : TestRun -> List TestRunAsset -> Html Msg
viewRunDetails run assets =
    Html.div []
        [ Html.h2 [ Html.Attributes.class "mdc-typography--headline5" ] [ Html.text "Run Details" ]
        , Html.div []
            [ Html.strong [] [ Html.text "Status: " ]
            , Html.text (statusToString run.status)
            ]
        , Html.div []
            [ Html.strong [] [ Html.text "Notes: " ]
            , Html.text run.notes
            ]
        , Html.div []
            [ Html.strong [] [ Html.text "Created: " ]
            , Html.text (formatTime run.createdAt)
            ]
        , case run.startedAt of
            Just startedAt ->
                Html.div []
                    [ Html.strong [] [ Html.text "Started: " ]
                    , Html.text (formatTime startedAt)
                    ]

            Nothing ->
                Html.text ""
        , case run.completedAt of
            Just completedAt ->
                Html.div []
                    [ Html.strong [] [ Html.text "Completed: " ]
                    , Html.text (formatTime completedAt)
                    ]

            Nothing ->
                Html.text ""
        , Html.h3 [ Html.Attributes.class "mdc-typography--headline6" ] [ Html.text "Assets" ]
        , if List.isEmpty assets then
            Html.div [] [ Html.text "No assets uploaded yet" ]

          else
            Html.div []
                (List.map (viewAsset run.id) assets)
        , Html.div
            [ Html.Attributes.style "margin-top" "20px" ]
            [ Html.text "Upload assets using the API or file upload form" ]
        ]


viewAsset : String -> TestRunAsset -> Html Msg
viewAsset runId asset =
    Html.div
        [ Html.Attributes.style "border" "1px solid #ddd"
        , Html.Attributes.style "border-radius" "4px"
        , Html.Attributes.style "padding" "16px"
        , Html.Attributes.style "margin" "10px 0"
        ]
        [ Html.div []
            [ Html.strong [] [ Html.text "Type: " ]
            , Html.text (assetTypeToString asset.assetType)
            ]
        , Html.div []
            [ Html.strong [] [ Html.text "Filename: " ]
            , Html.text asset.filename
            ]
        , Html.div []
            [ Html.strong [] [ Html.text "Description: " ]
            , Html.text asset.description
            ]
        , Html.div
            [ Html.Attributes.style "margin-top" "10px" ]
            [ Html.a
                [ Html.Attributes.href (API.getAssetDownloadUrl runId asset.id)
                , Html.Attributes.target "_blank"
                ]
                [ Html.text "Download" ]
            , Html.text " | "
            , Html.button
                [ Html.Events.onClick (DeleteAsset runId asset.id)
                , Html.Attributes.class "mdc-button"
                ]
                [ Html.text "Delete" ]
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
    Components.viewDialogOverlay "Create Test Run"
        [ Components.viewFormField "Notes"
            [ Html.Attributes.type_ "text"
            , Html.Attributes.value dialog.notes
            , Html.Events.onInput SetCreateNotes
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


viewCompleteDialog : CompleteDialogState -> Html Msg
viewCompleteDialog dialog =
    Components.viewDialogOverlay "Complete Test Run"
        [ Html.div
            [ Html.Attributes.style "margin-bottom" "20px"
            ]
            [ Html.label
                [ Html.Attributes.style "display" "block"
                , Html.Attributes.style "margin-bottom" "8px"
                , Html.Attributes.style "font-weight" "500"
                , Html.Attributes.style "color" "#333"
                ]
                [ Html.text "Status" ]
            , Html.select
                [ Html.Events.onInput SetCompleteStatus
                , Html.Attributes.style "width" "100%"
                , Html.Attributes.style "padding" "10px"
                , Html.Attributes.style "border" "1px solid #ddd"
                , Html.Attributes.style "border-radius" "4px"
                , Html.Attributes.style "font-size" "14px"
                , Html.Attributes.style "box-sizing" "border-box"
                ]
                [ Html.option
                    [ Html.Attributes.value "passed"
                    , Html.Attributes.selected (dialog.status == Types.Passed)
                    ]
                    [ Html.text "Passed" ]
                , Html.option
                    [ Html.Attributes.value "failed"
                    , Html.Attributes.selected (dialog.status == Types.Failed)
                    ]
                    [ Html.text "Failed" ]
                , Html.option
                    [ Html.Attributes.value "skipped"
                    , Html.Attributes.selected (dialog.status == Types.Skipped)
                    ]
                    [ Html.text "Skipped" ]
                ]
            ]
        , Components.viewFormField "Notes"
            [ Html.Attributes.type_ "text"
            , Html.Attributes.value dialog.notes
            , Html.Events.onInput SetCompleteNotes
            ]
        ]
        [ Html.button
            [ Html.Events.onClick CloseCompleteDialog
            , Html.Attributes.class "mdc-button"
            ]
            [ Html.text "Cancel" ]
        , Html.button
            [ Html.Events.onClick SubmitComplete
            , Html.Attributes.class "mdc-button mdc-button--raised"
            ]
            [ Html.text "Complete" ]
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


statusToString : TestRunStatus -> String
statusToString status =
    case status of
        Types.Pending ->
            "Pending"

        Types.Running ->
            "Running"

        Types.Passed ->
            "Passed"

        Types.Failed ->
            "Failed"

        Types.Skipped ->
            "Skipped"


stringToStatus : String -> TestRunStatus
stringToStatus str =
    case str of
        "passed" ->
            Types.Passed

        "failed" ->
            Types.Failed

        "skipped" ->
            Types.Skipped

        "running" ->
            Types.Running

        _ ->
            Types.Pending


assetTypeToString : AssetType -> String
assetTypeToString assetType =
    case assetType of
        Types.Image ->
            "Image"

        Types.Video ->
            "Video"

        Types.Binary ->
            "Binary"

        Types.Document ->
            "Document"
