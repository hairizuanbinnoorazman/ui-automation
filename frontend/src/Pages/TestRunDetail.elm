module Pages.TestRunDetail exposing (Model, Msg, init, update, view)

import API
import Dict exposing (Dict)
import File exposing (File)
import Html exposing (Html)
import Html.Attributes
import Html.Events
import Http
import Json.Decode as Decode
import Time
import Types exposing (CompleteTestRunInput, TestProcedure, TestRun, TestRunAsset, TestRunStepNote, TestRunStatus)



-- MODEL


type alias CompleteDialogState =
    { status : TestRunStatus
    , notes : String
    }


type alias Model =
    { runId : String
    , run : Maybe TestRun
    , procedure : Maybe TestProcedure
    , stepNotes : Dict Int String
    , savedStepNotes : Dict Int String
    , stepAssets : Dict Int (List TestRunAsset)
    , allAssets : List TestRunAsset
    , loading : Bool
    , error : Maybe String
    , completeDialog : Maybe CompleteDialogState
    }


init : String -> ( Model, Cmd Msg )
init runId =
    ( { runId = runId
      , run = Nothing
      , procedure = Nothing
      , stepNotes = Dict.empty
      , savedStepNotes = Dict.empty
      , stepAssets = Dict.empty
      , allAssets = []
      , loading = True
      , error = Nothing
      , completeDialog = Nothing
      }
    , Cmd.batch
        [ API.getTestRun runId RunResponse
        , API.getStepNotes runId StepNotesResponse
        , API.getTestRunAssets runId AssetsResponse
        , API.getRunProcedure runId ProcedureResponse
        ]
    )



-- UPDATE


type Msg
    = RunResponse (Result Http.Error TestRun)
    | ProcedureResponse (Result Http.Error TestProcedure)
    | StepNotesResponse (Result Http.Error (List TestRunStepNote))
    | AssetsResponse (Result Http.Error (List TestRunAsset))
    | StartRun
    | StartRunResponse (Result Http.Error TestRun)
    | OpenCompleteDialog
    | CloseCompleteDialog
    | SetCompleteStatus String
    | SetCompleteNotes String
    | SubmitComplete
    | CompleteResponse (Result Http.Error TestRun)
    | SetStepNote Int String
    | SaveAllNotes
    | StepNoteSaved Int (Result Http.Error TestRunStepNote)
    | FileSelected Int File
    | UploadAssetResponse Int (Result Http.Error TestRunAsset)


update : Msg -> Model -> ( Model, Cmd Msg )
update msg model =
    case msg of
        RunResponse (Ok run) ->
            ( { model | run = Just run, loading = False }
            , Cmd.none
            )

        RunResponse (Err error) ->
            ( { model | loading = False, error = Just ("Failed to load test run: " ++ httpErrorToString error) }
            , Cmd.none
            )

        ProcedureResponse (Ok proc) ->
            ( { model | procedure = Just proc }
            , Cmd.none
            )

        ProcedureResponse (Err error) ->
            ( { model | error = Just ("Failed to load procedure: " ++ httpErrorToString error) }
            , Cmd.none
            )

        StepNotesResponse (Ok notes) ->
            let
                savedNotes =
                    List.foldl
                        (\note acc -> Dict.insert note.stepIndex note.notes acc)
                        Dict.empty
                        notes
            in
            ( { model
                | savedStepNotes = savedNotes
                , stepNotes = savedNotes
              }
            , Cmd.none
            )

        StepNotesResponse (Err error) ->
            ( { model | error = Just ("Failed to load step notes: " ++ httpErrorToString error) }
            , Cmd.none
            )

        AssetsResponse (Ok assets) ->
            let
                byStep =
                    List.foldl
                        (\asset acc ->
                            case asset.stepIndex of
                                Just idx ->
                                    Dict.update idx
                                        (\existing ->
                                            case existing of
                                                Just list ->
                                                    Just (list ++ [ asset ])

                                                Nothing ->
                                                    Just [ asset ]
                                        )
                                        acc

                                Nothing ->
                                    acc
                        )
                        Dict.empty
                        assets
            in
            ( { model | allAssets = assets, stepAssets = byStep }
            , Cmd.none
            )

        AssetsResponse (Err error) ->
            ( { model | error = Just ("Failed to load assets: " ++ httpErrorToString error) }
            , Cmd.none
            )

        StartRun ->
            ( { model | loading = True }
            , API.startTestRun model.runId StartRunResponse
            )

        StartRunResponse (Ok run) ->
            ( { model | run = Just run, loading = False }
            , Cmd.none
            )

        StartRunResponse (Err error) ->
            ( { model | loading = False, error = Just (httpErrorToString error) }
            , Cmd.none
            )

        OpenCompleteDialog ->
            ( { model
                | completeDialog =
                    Just
                        { status = Types.Passed
                        , notes = ""
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
                    ( { model | completeDialog = Just { dialog | status = stringToStatus statusStr } }
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
                        model.runId
                        { status = dialog.status, notes = dialog.notes }
                        CompleteResponse
                    )

                Nothing ->
                    ( model, Cmd.none )

        CompleteResponse (Ok run) ->
            ( { model | run = Just run, loading = False, completeDialog = Nothing }
            , Cmd.none
            )

        CompleteResponse (Err error) ->
            ( { model | loading = False, error = Just (httpErrorToString error) }
            , Cmd.none
            )

        SetStepNote stepIndex notes ->
            ( { model | stepNotes = Dict.insert stepIndex notes model.stepNotes }
            , Cmd.none
            )

        SaveAllNotes ->
            case model.procedure of
                Just proc ->
                    let
                        cmds =
                            List.indexedMap
                                (\idx _ ->
                                    API.setStepNote
                                        model.runId
                                        idx
                                        (Dict.get idx model.stepNotes |> Maybe.withDefault "")
                                        (StepNoteSaved idx)
                                )
                                proc.steps
                    in
                    ( model, Cmd.batch cmds )

                Nothing ->
                    ( model, Cmd.none )

        StepNoteSaved stepIndex (Ok note) ->
            ( { model | savedStepNotes = Dict.insert stepIndex note.notes model.savedStepNotes }
            , Cmd.none
            )

        StepNoteSaved _ (Err error) ->
            ( { model | error = Just (httpErrorToString error) }
            , Cmd.none
            )

        FileSelected stepIndex file ->
            ( model
            , API.uploadStepAsset model.runId stepIndex file (UploadAssetResponse stepIndex)
            )

        UploadAssetResponse stepIndex (Ok asset) ->
            let
                updatedAssets =
                    Dict.update stepIndex
                        (\existing ->
                            case existing of
                                Just list ->
                                    Just (list ++ [ asset ])

                                Nothing ->
                                    Just [ asset ]
                        )
                        model.stepAssets
            in
            ( { model | stepAssets = updatedAssets }
            , Cmd.none
            )

        UploadAssetResponse _ (Err error) ->
            ( { model | error = Just (httpErrorToString error) }
            , Cmd.none
            )



-- VIEW


view : Model -> Html Msg
view model =
    Html.div []
        [ case model.error of
            Just err ->
                Html.div
                    [ Html.Attributes.style "color" "red"
                    , Html.Attributes.style "margin-bottom" "20px"
                    ]
                    [ Html.text err ]

            Nothing ->
                Html.text ""
        , case model.run of
            Just _ ->
                viewRunHeader model

            Nothing ->
                if model.loading then
                    Html.div [] [ Html.text "Loading..." ]

                else
                    Html.div [] [ Html.text "Run not found" ]
        , case ( model.run, model.procedure ) of
            ( Just _, Just procedure ) ->
                viewSteps model procedure

            _ ->
                Html.text ""
        , case model.completeDialog of
            Just dialog ->
                viewCompleteDialog dialog

            Nothing ->
                Html.text ""
        ]


viewRunHeader : Model -> Html Msg
viewRunHeader model =
    case model.run of
        Nothing ->
            Html.text ""

        Just run ->
            Html.div
                [ Html.Attributes.style "margin-bottom" "24px"
                , Html.Attributes.style "padding" "16px"
                , Html.Attributes.style "border" "1px solid #ddd"
                , Html.Attributes.style "border-radius" "8px"
                ]
                [ Html.div
                    [ Html.Attributes.class "page-header" ]
                    [ Html.h1 [ Html.Attributes.class "mdc-typography--headline3" ] [ Html.text "Test Run Detail" ]
                    , Html.div
                        [ Html.Attributes.style "display" "flex"
                        , Html.Attributes.style "gap" "8px"
                        , Html.Attributes.style "align-items" "center"
                        ]
                        [ Html.button
                            [ Html.Events.onClick SaveAllNotes
                            , Html.Attributes.class "mdc-button mdc-button--raised"
                            ]
                            [ Html.text "Save Notes" ]
                        , if run.status == Types.Pending then
                            Html.button
                                [ Html.Events.onClick StartRun
                                , Html.Attributes.class "mdc-button mdc-button--raised"
                                ]
                                [ Html.text "Start" ]

                          else if run.status == Types.Running then
                            Html.button
                                [ Html.Events.onClick OpenCompleteDialog
                                , Html.Attributes.class "mdc-button mdc-button--raised"
                                ]
                                [ Html.text "Complete" ]

                          else
                            Html.text ""
                        , Html.a
                            [ Html.Attributes.href ("/api/v1/runs/" ++ run.id ++ "/guide")
                            , Html.Attributes.download ""
                            , Html.Attributes.class "mdc-button mdc-button--outlined"
                            ]
                            [ Html.text "Generate Guide" ]
                        ]
                    ]
                , Html.div
                    [ Html.Attributes.style "display" "flex"
                    , Html.Attributes.style "gap" "24px"
                    , Html.Attributes.style "flex-wrap" "wrap"
                    ]
                    [ Html.div []
                        [ Html.strong [] [ Html.text "Status: " ]
                        , Html.span
                            [ Html.Attributes.style "font-weight" "bold"
                            , Html.Attributes.style "color" (statusColor run.status)
                            ]
                            [ Html.text (statusToString run.status) ]
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
                    ]
                ]


viewSteps : Model -> TestProcedure -> Html Msg
viewSteps model procedure =
    Html.div []
        [ Html.h2 [ Html.Attributes.class "mdc-typography--headline5" ] [ Html.text "Steps" ]
        , Html.div []
            (List.indexedMap (viewStep model) procedure.steps)
        ]


viewStep : Model -> Int -> Types.TestStep -> Html Msg
viewStep model stepIndex step =
    let
        currentNotes =
            Dict.get stepIndex model.stepNotes
                |> Maybe.withDefault ""

        stepAssets =
            Dict.get stepIndex model.stepAssets
                |> Maybe.withDefault []
    in
    Html.div
        [ Html.Attributes.style "border" "1px solid #ddd"
        , Html.Attributes.style "border-radius" "8px"
        , Html.Attributes.style "padding" "16px"
        , Html.Attributes.style "margin-bottom" "16px"
        ]
        [ Html.div
            [ Html.Attributes.style "display" "flex"
            , Html.Attributes.style "justify-content" "space-between"
            , Html.Attributes.style "align-items" "center"
            , Html.Attributes.style "margin-bottom" "8px"
            ]
            [ Html.h3
                [ Html.Attributes.class "mdc-typography--headline6"
                , Html.Attributes.style "margin" "0"
                ]
                [ Html.text ("Step " ++ String.fromInt (stepIndex + 1) ++ ": " ++ step.name) ]
            ]
        , Html.p
            [ Html.Attributes.style "margin-bottom" "12px"
            , Html.Attributes.style "color" "#555"
            ]
            [ Html.text step.instructions ]
        , if not (List.isEmpty step.imagePaths) then
            Html.div
                [ Html.Attributes.style "margin-bottom" "12px" ]
                [ Html.strong [] [ Html.text "Reference images:" ]
                , Html.div
                    [ Html.Attributes.style "display" "flex"
                    , Html.Attributes.style "gap" "8px"
                    , Html.Attributes.style "flex-wrap" "wrap"
                    , Html.Attributes.style "margin-top" "8px"
                    ]
                    (List.map viewStepImage step.imagePaths)
                ]

          else
            Html.text ""
        , Html.div
            [ Html.Attributes.style "margin-bottom" "12px" ]
            [ Html.label
                [ Html.Attributes.style "display" "block"
                , Html.Attributes.style "margin-bottom" "4px"
                , Html.Attributes.style "font-weight" "bold"
                ]
                [ Html.text "Notes:" ]
            , Html.textarea
                [ Html.Attributes.value currentNotes
                , Html.Events.onInput (SetStepNote stepIndex)
                , Html.Attributes.style "width" "100%"
                , Html.Attributes.style "min-height" "80px"
                , Html.Attributes.style "padding" "8px"
                , Html.Attributes.style "box-sizing" "border-box"
                , Html.Attributes.style "border" "1px solid #ccc"
                , Html.Attributes.style "border-radius" "4px"
                , Html.Attributes.style "font-family" "inherit"
                , Html.Attributes.style "font-size" "14px"
                ]
                []
            ]
        , Html.div []
            [ Html.strong [] [ Html.text "Step Images:" ]
            , if List.isEmpty stepAssets then
                Html.p
                    [ Html.Attributes.style "color" "#999"
                    , Html.Attributes.style "font-size" "14px"
                    ]
                    [ Html.text "No images uploaded for this step" ]

              else
                Html.div
                    [ Html.Attributes.style "display" "flex"
                    , Html.Attributes.style "gap" "8px"
                    , Html.Attributes.style "flex-wrap" "wrap"
                    , Html.Attributes.style "margin-top" "8px"
                    ]
                    (List.map viewUploadedAsset stepAssets)
            , Html.div
                [ Html.Attributes.style "margin-top" "8px" ]
                [ Html.label
                    [ Html.Attributes.style "display" "inline-block"
                    , Html.Attributes.style "cursor" "pointer"
                    , Html.Attributes.style "padding" "6px 12px"
                    , Html.Attributes.style "border" "1px solid #6200ee"
                    , Html.Attributes.style "border-radius" "4px"
                    , Html.Attributes.style "color" "#6200ee"
                    , Html.Attributes.style "font-size" "14px"
                    ]
                    [ Html.text "Upload Image"
                    , Html.input
                        [ Html.Attributes.type_ "file"
                        , Html.Attributes.accept "image/*"
                        , Html.Attributes.style "display" "none"
                        , Html.Events.on "change" (Decode.map (FileSelected stepIndex) fileDecoder)
                        ]
                        []
                    ]
                ]
            ]
        ]


viewStepImage : String -> Html Msg
viewStepImage imagePath =
    let
        fullPath =
            "/uploads/" ++ imagePath
    in
    Html.a
        [ Html.Attributes.href fullPath
        , Html.Attributes.target "_blank"
        ]
        [ Html.img
            [ Html.Attributes.src fullPath
            , Html.Attributes.style "max-width" "120px"
            , Html.Attributes.style "max-height" "120px"
            , Html.Attributes.style "object-fit" "cover"
            , Html.Attributes.style "border-radius" "4px"
            , Html.Attributes.style "border" "1px solid #ddd"
            ]
            []
        ]


viewUploadedAsset : TestRunAsset -> Html Msg
viewUploadedAsset asset =
    Html.a
        [ Html.Attributes.href ("/api/v1/runs/" ++ asset.testRunId ++ "/assets/" ++ asset.id)
        , Html.Attributes.target "_blank"
        , Html.Attributes.style "display" "block"
        , Html.Attributes.style "font-size" "14px"
        , Html.Attributes.style "color" "#6200ee"
        ]
        [ Html.text asset.filename ]


viewCompleteDialog : CompleteDialogState -> Html Msg
viewCompleteDialog dialog =
    Html.div
        [ Html.Attributes.style "position" "fixed"
        , Html.Attributes.style "top" "0"
        , Html.Attributes.style "left" "0"
        , Html.Attributes.style "width" "100%"
        , Html.Attributes.style "height" "100%"
        , Html.Attributes.style "background" "rgba(0,0,0,0.5)"
        , Html.Attributes.style "display" "flex"
        , Html.Attributes.style "align-items" "center"
        , Html.Attributes.style "justify-content" "center"
        , Html.Attributes.style "z-index" "1000"
        ]
        [ Html.div
            [ Html.Attributes.style "background" "white"
            , Html.Attributes.style "border-radius" "8px"
            , Html.Attributes.style "padding" "24px"
            , Html.Attributes.style "min-width" "360px"
            , Html.Attributes.style "max-width" "500px"
            ]
            [ Html.h2
                [ Html.Attributes.class "mdc-typography--headline6"
                , Html.Attributes.style "margin-top" "0"
                ]
                [ Html.text "Complete Test Run" ]
            , Html.div
                [ Html.Attributes.style "margin-bottom" "16px" ]
                [ Html.label
                    [ Html.Attributes.style "display" "block"
                    , Html.Attributes.style "margin-bottom" "4px"
                    ]
                    [ Html.text "Status" ]
                , Html.select
                    [ Html.Events.onInput SetCompleteStatus
                    , Html.Attributes.style "width" "100%"
                    , Html.Attributes.style "padding" "8px"
                    , Html.Attributes.style "border" "1px solid #ccc"
                    , Html.Attributes.style "border-radius" "4px"
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
            , Html.div
                [ Html.Attributes.style "margin-bottom" "16px" ]
                [ Html.label
                    [ Html.Attributes.style "display" "block"
                    , Html.Attributes.style "margin-bottom" "4px"
                    ]
                    [ Html.text "Notes" ]
                , Html.textarea
                    [ Html.Attributes.value dialog.notes
                    , Html.Events.onInput SetCompleteNotes
                    , Html.Attributes.style "width" "100%"
                    , Html.Attributes.style "min-height" "80px"
                    , Html.Attributes.style "padding" "8px"
                    , Html.Attributes.style "box-sizing" "border-box"
                    , Html.Attributes.style "border" "1px solid #ccc"
                    , Html.Attributes.style "border-radius" "4px"
                    , Html.Attributes.style "font-family" "inherit"
                    ]
                    []
                ]
            , Html.div
                [ Html.Attributes.style "display" "flex"
                , Html.Attributes.style "justify-content" "flex-end"
                , Html.Attributes.style "gap" "8px"
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
            ]
        ]



-- HELPERS


fileDecoder : Decode.Decoder File
fileDecoder =
    Decode.at [ "target", "files", "0" ] File.decoder


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


statusColor : TestRunStatus -> String
statusColor status =
    case status of
        Types.Pending ->
            "#f57c00"

        Types.Running ->
            "#1976d2"

        Types.Passed ->
            "#388e3c"

        Types.Failed ->
            "#d32f2f"

        Types.Skipped ->
            "#757575"


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
