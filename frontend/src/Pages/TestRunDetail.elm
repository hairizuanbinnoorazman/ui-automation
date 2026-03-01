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
import Types exposing (CompleteTestRunInput, CreateIssueLinkInput, ExternalIssue, Integration, IntegrationListResponse, IssueLink, LinkExistingIssueInput, TestProcedure, TestRun, TestRunAsset, TestRunStepNote, TestRunStatus, User, UserListResponse)



-- MODEL


type alias CompleteDialogState =
    { status : TestRunStatus
    , notes : String
    }


type alias AssignDialogState =
    { searchQuery : String
    , searchResults : List User
    , selectedUser : Maybe User
    , loading : Bool
    }


type alias CreateIssueDialogState =
    { integrationId : String
    , title : String
    , description : String
    , projectKey : String
    , issueType : String
    , repository : String
    }


type alias LinkIssueDialogState =
    { integrationId : String
    , searchQuery : String
    , searchResults : List ExternalIssue
    , selectedIssue : Maybe ExternalIssue
    , loading : Bool
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
    , assignDialog : Maybe AssignDialogState
    , assignedUser : Maybe User
    , issueLinks : List IssueLink
    , issuesLoading : Bool
    , integrations : List Integration
    , createIssueDialog : Maybe CreateIssueDialogState
    , linkIssueDialog : Maybe LinkIssueDialogState
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
      , assignDialog = Nothing
      , assignedUser = Nothing
      , issueLinks = []
      , issuesLoading = True
      , integrations = []
      , createIssueDialog = Nothing
      , linkIssueDialog = Nothing
      }
    , Cmd.batch
        [ API.getTestRun runId RunResponse
        , API.getStepNotes runId StepNotesResponse
        , API.getTestRunAssets runId AssetsResponse
        , API.getRunProcedure runId ProcedureResponse
        , API.getIssueLinks runId IssueLinksResponse
        , API.getIntegrations IntegrationsResponse
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
    | OpenAssignDialog
    | CloseAssignDialog
    | SetAssignSearchQuery String
    | SearchUsersResponse (Result Http.Error UserListResponse)
    | SelectAssignUser User
    | SubmitAssign
    | AssignResponse (Result Http.Error TestRun)
    | UnassignUser
    | UnassignResponse (Result Http.Error TestRun)
    | AssignedUserResponse (Result Http.Error User)
    | IssueLinksResponse (Result Http.Error (List IssueLink))
    | IntegrationsResponse (Result Http.Error IntegrationListResponse)
    | OpenCreateIssueDialog
    | CloseCreateIssueDialog
    | SetCreateIssueIntegration String
    | SetCreateIssueTitle String
    | SetCreateIssueDescription String
    | SetCreateIssueProjectKey String
    | SetCreateIssueType String
    | SetCreateIssueRepository String
    | SubmitCreateIssue
    | CreateIssueResponse (Result Http.Error IssueLink)
    | OpenLinkIssueDialog
    | CloseLinkIssueDialog
    | SetLinkIssueIntegration String
    | SetLinkIssueSearchQuery String
    | SearchExternalIssues
    | SearchExternalIssuesResponse (Result Http.Error (List ExternalIssue))
    | SelectExternalIssue ExternalIssue
    | SubmitLinkIssue
    | LinkIssueResponse (Result Http.Error IssueLink)
    | UnlinkIssue String
    | UnlinkIssueResponse (Result Http.Error ())
    | ResolveIssue String
    | ResolveIssueResponse (Result Http.Error IssueLink)
    | SyncIssue String
    | SyncIssueResponse (Result Http.Error IssueLink)


update : Msg -> Model -> ( Model, Cmd Msg )
update msg model =
    case msg of
        RunResponse (Ok run) ->
            let
                fetchAssignedUser =
                    case run.assignedTo of
                        Just userId ->
                            API.getUserById userId AssignedUserResponse

                        Nothing ->
                            Cmd.none
            in
            ( { model | run = Just run, loading = False }
            , fetchAssignedUser
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

        OpenAssignDialog ->
            ( { model
                | assignDialog =
                    Just
                        { searchQuery = ""
                        , searchResults = []
                        , selectedUser = Nothing
                        , loading = False
                        }
              }
            , Cmd.none
            )

        CloseAssignDialog ->
            ( { model | assignDialog = Nothing }
            , Cmd.none
            )

        SetAssignSearchQuery query ->
            case model.assignDialog of
                Just dialog ->
                    let
                        updatedDialog =
                            { dialog | searchQuery = query, loading = True }

                        cmd =
                            if String.length query >= 2 then
                                API.searchUsers query SearchUsersResponse

                            else
                                Cmd.none
                    in
                    ( { model | assignDialog = Just updatedDialog }
                    , cmd
                    )

                Nothing ->
                    ( model, Cmd.none )

        SearchUsersResponse (Ok response) ->
            case model.assignDialog of
                Just dialog ->
                    ( { model | assignDialog = Just { dialog | searchResults = response.users, loading = False } }
                    , Cmd.none
                    )

                Nothing ->
                    ( model, Cmd.none )

        SearchUsersResponse (Err _) ->
            case model.assignDialog of
                Just dialog ->
                    ( { model | assignDialog = Just { dialog | loading = False } }
                    , Cmd.none
                    )

                Nothing ->
                    ( model, Cmd.none )

        SelectAssignUser selectedUser ->
            case model.assignDialog of
                Just dialog ->
                    ( { model | assignDialog = Just { dialog | selectedUser = Just selectedUser } }
                    , Cmd.none
                    )

                Nothing ->
                    ( model, Cmd.none )

        SubmitAssign ->
            case model.assignDialog of
                Just dialog ->
                    case dialog.selectedUser of
                        Just selectedUser ->
                            ( { model | loading = True }
                            , API.assignTestRunUser model.runId selectedUser.id AssignResponse
                            )

                        Nothing ->
                            ( model, Cmd.none )

                Nothing ->
                    ( model, Cmd.none )

        AssignResponse (Ok run) ->
            let
                fetchAssignedUser =
                    case run.assignedTo of
                        Just userId ->
                            API.getUserById userId AssignedUserResponse

                        Nothing ->
                            Cmd.none
            in
            ( { model | run = Just run, loading = False, assignDialog = Nothing }
            , fetchAssignedUser
            )

        AssignResponse (Err error) ->
            ( { model | loading = False, error = Just (httpErrorToString error) }
            , Cmd.none
            )

        UnassignUser ->
            ( { model | loading = True }
            , API.unassignTestRunUser model.runId UnassignResponse
            )

        UnassignResponse (Ok run) ->
            ( { model | run = Just run, loading = False, assignedUser = Nothing }
            , Cmd.none
            )

        UnassignResponse (Err error) ->
            ( { model | loading = False, error = Just (httpErrorToString error) }
            , Cmd.none
            )

        AssignedUserResponse (Ok fetchedUser) ->
            ( { model | assignedUser = Just fetchedUser }
            , Cmd.none
            )

        AssignedUserResponse (Err _) ->
            ( model, Cmd.none )

        IssueLinksResponse (Ok links) ->
            ( { model | issueLinks = links, issuesLoading = False }
            , Cmd.none
            )

        IssueLinksResponse (Err error) ->
            ( { model | issuesLoading = False, error = Just ("Failed to load issues: " ++ httpErrorToString error) }
            , Cmd.none
            )

        IntegrationsResponse (Ok response) ->
            ( { model | integrations = response.items }
            , Cmd.none
            )

        IntegrationsResponse (Err _) ->
            ( model, Cmd.none )

        OpenCreateIssueDialog ->
            let
                defaultIntegrationId =
                    List.head model.integrations
                        |> Maybe.map .id
                        |> Maybe.withDefault ""

                defaultProvider =
                    List.head model.integrations
                        |> Maybe.map .provider
                        |> Maybe.withDefault ""
            in
            ( { model
                | createIssueDialog =
                    Just
                        { integrationId = defaultIntegrationId
                        , title = ""
                        , description = ""
                        , projectKey = ""
                        , issueType =
                            if defaultProvider == "jira" then
                                "Bug"

                            else
                                ""
                        , repository = ""
                        }
              }
            , Cmd.none
            )

        CloseCreateIssueDialog ->
            ( { model | createIssueDialog = Nothing }
            , Cmd.none
            )

        SetCreateIssueIntegration integrationId ->
            case model.createIssueDialog of
                Just dialog ->
                    ( { model | createIssueDialog = Just { dialog | integrationId = integrationId } }
                    , Cmd.none
                    )

                Nothing ->
                    ( model, Cmd.none )

        SetCreateIssueTitle title ->
            case model.createIssueDialog of
                Just dialog ->
                    ( { model | createIssueDialog = Just { dialog | title = title } }
                    , Cmd.none
                    )

                Nothing ->
                    ( model, Cmd.none )

        SetCreateIssueDescription description ->
            case model.createIssueDialog of
                Just dialog ->
                    ( { model | createIssueDialog = Just { dialog | description = description } }
                    , Cmd.none
                    )

                Nothing ->
                    ( model, Cmd.none )

        SetCreateIssueProjectKey projectKey ->
            case model.createIssueDialog of
                Just dialog ->
                    ( { model | createIssueDialog = Just { dialog | projectKey = projectKey } }
                    , Cmd.none
                    )

                Nothing ->
                    ( model, Cmd.none )

        SetCreateIssueType issueType ->
            case model.createIssueDialog of
                Just dialog ->
                    ( { model | createIssueDialog = Just { dialog | issueType = issueType } }
                    , Cmd.none
                    )

                Nothing ->
                    ( model, Cmd.none )

        SetCreateIssueRepository repository ->
            case model.createIssueDialog of
                Just dialog ->
                    ( { model | createIssueDialog = Just { dialog | repository = repository } }
                    , Cmd.none
                    )

                Nothing ->
                    ( model, Cmd.none )

        SubmitCreateIssue ->
            case model.createIssueDialog of
                Just dialog ->
                    ( { model | issuesLoading = True }
                    , API.createAndLinkIssue model.runId
                        { integrationId = dialog.integrationId
                        , title = dialog.title
                        , description = dialog.description
                        , projectKey = dialog.projectKey
                        , issueType = dialog.issueType
                        , repository = dialog.repository
                        , labels = []
                        }
                        CreateIssueResponse
                    )

                Nothing ->
                    ( model, Cmd.none )

        CreateIssueResponse (Ok link) ->
            ( { model
                | issuesLoading = False
                , createIssueDialog = Nothing
                , issueLinks = model.issueLinks ++ [ link ]
              }
            , Cmd.none
            )

        CreateIssueResponse (Err error) ->
            ( { model | issuesLoading = False, error = Just (httpErrorToString error) }
            , Cmd.none
            )

        OpenLinkIssueDialog ->
            let
                defaultIntegrationId =
                    List.head model.integrations
                        |> Maybe.map .id
                        |> Maybe.withDefault ""
            in
            ( { model
                | linkIssueDialog =
                    Just
                        { integrationId = defaultIntegrationId
                        , searchQuery = ""
                        , searchResults = []
                        , selectedIssue = Nothing
                        , loading = False
                        }
              }
            , Cmd.none
            )

        CloseLinkIssueDialog ->
            ( { model | linkIssueDialog = Nothing }
            , Cmd.none
            )

        SetLinkIssueIntegration integrationId ->
            case model.linkIssueDialog of
                Just dialog ->
                    ( { model | linkIssueDialog = Just { dialog | integrationId = integrationId, searchResults = [], selectedIssue = Nothing } }
                    , Cmd.none
                    )

                Nothing ->
                    ( model, Cmd.none )

        SetLinkIssueSearchQuery query ->
            case model.linkIssueDialog of
                Just dialog ->
                    ( { model | linkIssueDialog = Just { dialog | searchQuery = query } }
                    , Cmd.none
                    )

                Nothing ->
                    ( model, Cmd.none )

        SearchExternalIssues ->
            case model.linkIssueDialog of
                Just dialog ->
                    if String.length dialog.searchQuery >= 2 then
                        ( { model | linkIssueDialog = Just { dialog | loading = True } }
                        , API.searchExternalIssues dialog.integrationId dialog.searchQuery SearchExternalIssuesResponse
                        )

                    else
                        ( model, Cmd.none )

                Nothing ->
                    ( model, Cmd.none )

        SearchExternalIssuesResponse (Ok results) ->
            case model.linkIssueDialog of
                Just dialog ->
                    ( { model | linkIssueDialog = Just { dialog | searchResults = results, loading = False } }
                    , Cmd.none
                    )

                Nothing ->
                    ( model, Cmd.none )

        SearchExternalIssuesResponse (Err _) ->
            case model.linkIssueDialog of
                Just dialog ->
                    ( { model | linkIssueDialog = Just { dialog | loading = False } }
                    , Cmd.none
                    )

                Nothing ->
                    ( model, Cmd.none )

        SelectExternalIssue issue ->
            case model.linkIssueDialog of
                Just dialog ->
                    ( { model | linkIssueDialog = Just { dialog | selectedIssue = Just issue } }
                    , Cmd.none
                    )

                Nothing ->
                    ( model, Cmd.none )

        SubmitLinkIssue ->
            case model.linkIssueDialog of
                Just dialog ->
                    case dialog.selectedIssue of
                        Just issue ->
                            ( { model | issuesLoading = True }
                            , API.linkExistingIssue model.runId
                                { integrationId = dialog.integrationId
                                , externalId = issue.externalId
                                }
                                LinkIssueResponse
                            )

                        Nothing ->
                            ( model, Cmd.none )

                Nothing ->
                    ( model, Cmd.none )

        LinkIssueResponse (Ok link) ->
            ( { model
                | issuesLoading = False
                , linkIssueDialog = Nothing
                , issueLinks = model.issueLinks ++ [ link ]
              }
            , Cmd.none
            )

        LinkIssueResponse (Err error) ->
            ( { model | issuesLoading = False, error = Just (httpErrorToString error) }
            , Cmd.none
            )

        UnlinkIssue linkId ->
            ( { model | issuesLoading = True }
            , API.unlinkIssue model.runId linkId UnlinkIssueResponse
            )

        UnlinkIssueResponse (Ok ()) ->
            ( { model | issuesLoading = False }
            , API.getIssueLinks model.runId IssueLinksResponse
            )

        UnlinkIssueResponse (Err error) ->
            ( { model | issuesLoading = False, error = Just (httpErrorToString error) }
            , Cmd.none
            )

        ResolveIssue linkId ->
            ( model
            , API.resolveLinkedIssue model.runId linkId ResolveIssueResponse
            )

        ResolveIssueResponse (Ok updatedLink) ->
            ( { model | issueLinks = updateIssueLink updatedLink model.issueLinks }
            , Cmd.none
            )

        ResolveIssueResponse (Err error) ->
            ( { model | error = Just (httpErrorToString error) }
            , Cmd.none
            )

        SyncIssue linkId ->
            ( model
            , API.syncIssueStatus model.runId linkId SyncIssueResponse
            )

        SyncIssueResponse (Ok updatedLink) ->
            ( { model | issueLinks = updateIssueLink updatedLink model.issueLinks }
            , Cmd.none
            )

        SyncIssueResponse (Err error) ->
            ( { model | error = Just (httpErrorToString error) }
            , Cmd.none
            )


updateIssueLink : IssueLink -> List IssueLink -> List IssueLink
updateIssueLink updatedLink links =
    List.map
        (\link ->
            if link.id == updatedLink.id then
                updatedLink

            else
                link
        )
        links



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
        , case model.run of
            Just _ ->
                viewIssuesSection model

            Nothing ->
                Html.text ""
        , case model.completeDialog of
            Just dialog ->
                viewCompleteDialog dialog

            Nothing ->
                Html.text ""
        , case model.assignDialog of
            Just dialog ->
                viewAssignDialog dialog

            Nothing ->
                Html.text ""
        , case model.createIssueDialog of
            Just dialog ->
                viewCreateIssueDialog model.integrations dialog

            Nothing ->
                Html.text ""
        , case model.linkIssueDialog of
            Just dialog ->
                viewLinkIssueDialog model.integrations dialog

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
                    , Html.Attributes.style "align-items" "center"
                    ]
                    [ Html.div []
                        [ Html.strong [] [ Html.text "Status: " ]
                        , Html.span
                            [ Html.Attributes.style "font-weight" "bold"
                            , Html.Attributes.style "color" (statusColor run.status)
                            ]
                            [ Html.text (statusToString run.status) ]
                        ]
                    , case model.procedure of
                        Just proc ->
                            Html.div []
                                [ Html.strong [] [ Html.text "Procedure Version: " ]
                                , Html.text ("v" ++ String.fromInt proc.version)
                                ]

                        Nothing ->
                            Html.text ""
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
                    , Html.div
                        [ Html.Attributes.style "display" "flex"
                        , Html.Attributes.style "align-items" "center"
                        , Html.Attributes.style "gap" "8px"
                        ]
                        [ Html.strong [] [ Html.text "Assigned To: " ]
                        , case model.assignedUser of
                            Just assignedUser ->
                                Html.span [] [ Html.text assignedUser.username ]

                            Nothing ->
                                case run.assignedTo of
                                    Just _ ->
                                        Html.span [ Html.Attributes.style "color" "#999" ] [ Html.text "Loading..." ]

                                    Nothing ->
                                        Html.span [ Html.Attributes.style "color" "#999" ] [ Html.text "Unassigned" ]
                        , Html.button
                            [ Html.Events.onClick OpenAssignDialog
                            , Html.Attributes.class "mdc-button mdc-button--outlined"
                            , Html.Attributes.style "font-size" "12px"
                            , Html.Attributes.style "padding" "2px 8px"
                            , Html.Attributes.style "min-height" "0"
                            ]
                            [ Html.text
                                (case run.assignedTo of
                                    Just _ ->
                                        "Reassign"

                                    Nothing ->
                                        "Assign"
                                )
                            ]
                        ]
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



viewAssignDialog : AssignDialogState -> Html Msg
viewAssignDialog dialog =
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
                [ Html.text "Assign User" ]
            , Html.div
                [ Html.Attributes.style "margin-bottom" "16px" ]
                [ Html.label
                    [ Html.Attributes.style "display" "block"
                    , Html.Attributes.style "margin-bottom" "4px"
                    ]
                    [ Html.text "Search users" ]
                , Html.input
                    [ Html.Attributes.type_ "text"
                    , Html.Attributes.placeholder "Type username or email..."
                    , Html.Attributes.value dialog.searchQuery
                    , Html.Events.onInput SetAssignSearchQuery
                    , Html.Attributes.style "width" "100%"
                    , Html.Attributes.style "padding" "8px"
                    , Html.Attributes.style "border" "1px solid #ccc"
                    , Html.Attributes.style "border-radius" "4px"
                    , Html.Attributes.style "box-sizing" "border-box"
                    , Html.Attributes.style "font-size" "14px"
                    ]
                    []
                ]
            , if not (List.isEmpty dialog.searchResults) then
                Html.div
                    [ Html.Attributes.style "max-height" "200px"
                    , Html.Attributes.style "overflow-y" "auto"
                    , Html.Attributes.style "border" "1px solid #ddd"
                    , Html.Attributes.style "border-radius" "4px"
                    , Html.Attributes.style "margin-bottom" "16px"
                    ]
                    (List.map (viewUserSearchResult dialog.selectedUser) dialog.searchResults)

              else if dialog.loading then
                Html.p
                    [ Html.Attributes.style "color" "#999"
                    , Html.Attributes.style "font-size" "14px"
                    , Html.Attributes.style "margin-bottom" "16px"
                    ]
                    [ Html.text "Searching..." ]

              else if String.length dialog.searchQuery >= 2 then
                Html.p
                    [ Html.Attributes.style "color" "#999"
                    , Html.Attributes.style "font-size" "14px"
                    , Html.Attributes.style "margin-bottom" "16px"
                    ]
                    [ Html.text "No users found" ]

              else
                Html.text ""
            , case dialog.selectedUser of
                Just selectedUser ->
                    Html.div
                        [ Html.Attributes.style "margin-bottom" "16px"
                        , Html.Attributes.style "padding" "8px"
                        , Html.Attributes.style "background" "#e8f5e9"
                        , Html.Attributes.style "border-radius" "4px"
                        ]
                        [ Html.text ("Selected: " ++ selectedUser.username ++ " (" ++ selectedUser.email ++ ")") ]

                Nothing ->
                    Html.text ""
            , Html.div
                [ Html.Attributes.style "display" "flex"
                , Html.Attributes.style "justify-content" "flex-end"
                , Html.Attributes.style "gap" "8px"
                ]
                [ Html.button
                    [ Html.Events.onClick UnassignUser
                    , Html.Attributes.class "mdc-button"
                    , Html.Attributes.style "color" "#d32f2f"
                    ]
                    [ Html.text "Unassign" ]
                , Html.button
                    [ Html.Events.onClick CloseAssignDialog
                    , Html.Attributes.class "mdc-button"
                    ]
                    [ Html.text "Cancel" ]
                , Html.button
                    [ Html.Events.onClick SubmitAssign
                    , Html.Attributes.class "mdc-button mdc-button--raised"
                    , Html.Attributes.disabled (dialog.selectedUser == Nothing)
                    ]
                    [ Html.text "Assign" ]
                ]
            ]
        ]


viewUserSearchResult : Maybe User -> User -> Html Msg
viewUserSearchResult selectedUser resultUser =
    let
        isSelected =
            case selectedUser of
                Just sel ->
                    sel.id == resultUser.id

                Nothing ->
                    False
    in
    Html.div
        [ Html.Events.onClick (SelectAssignUser resultUser)
        , Html.Attributes.style "padding" "8px 12px"
        , Html.Attributes.style "cursor" "pointer"
        , Html.Attributes.style "border-bottom" "1px solid #eee"
        , Html.Attributes.style "background"
            (if isSelected then
                "#e3f2fd"

             else
                "white"
            )
        ]
        [ Html.div
            [ Html.Attributes.style "font-weight" "bold"
            , Html.Attributes.style "font-size" "14px"
            ]
            [ Html.text resultUser.username ]
        , Html.div
            [ Html.Attributes.style "font-size" "12px"
            , Html.Attributes.style "color" "#666"
            ]
            [ Html.text resultUser.email ]
        ]



-- Issues Views


viewIssuesSection : Model -> Html Msg
viewIssuesSection model =
    Html.div
        [ Html.Attributes.style "margin-top" "32px" ]
        [ Html.div
            [ Html.Attributes.style "display" "flex"
            , Html.Attributes.style "justify-content" "space-between"
            , Html.Attributes.style "align-items" "center"
            , Html.Attributes.style "margin-bottom" "16px"
            ]
            [ Html.h2 [ Html.Attributes.class "mdc-typography--headline5" ] [ Html.text "Issues" ]
            , Html.div
                [ Html.Attributes.style "display" "flex"
                , Html.Attributes.style "gap" "8px"
                ]
                [ Html.button
                    [ Html.Events.onClick OpenCreateIssueDialog
                    , Html.Attributes.class "mdc-button mdc-button--raised"
                    , Html.Attributes.disabled (List.isEmpty model.integrations)
                    ]
                    [ Html.text "Create Issue" ]
                , Html.button
                    [ Html.Events.onClick OpenLinkIssueDialog
                    , Html.Attributes.class "mdc-button mdc-button--outlined"
                    , Html.Attributes.disabled (List.isEmpty model.integrations)
                    ]
                    [ Html.text "Link Existing" ]
                ]
            ]
        , if List.isEmpty model.integrations then
            Html.div
                [ Html.Attributes.style "color" "#999"
                , Html.Attributes.style "padding" "16px"
                , Html.Attributes.style "font-size" "14px"
                ]
                [ Html.text "No integrations configured. Go to Account Management to connect an integration." ]

          else if model.issuesLoading && List.isEmpty model.issueLinks then
            Html.div [] [ Html.text "Loading issues..." ]

          else if List.isEmpty model.issueLinks then
            Html.div
                [ Html.Attributes.style "color" "#666"
                , Html.Attributes.style "padding" "20px"
                ]
                [ Html.text "No issues linked to this test run." ]

          else
            viewIssueLinksTable model.issueLinks
        ]


viewIssueLinksTable : List IssueLink -> Html Msg
viewIssueLinksTable links =
    Html.table
        [ Html.Attributes.class "mdc-data-table__table"
        , Html.Attributes.style "width" "100%"
        , Html.Attributes.style "border-collapse" "collapse"
        ]
        [ Html.thead []
            [ Html.tr []
                [ Html.th [ Html.Attributes.style "text-align" "left", Html.Attributes.style "padding" "12px" ] [ Html.text "Provider" ]
                , Html.th [ Html.Attributes.style "text-align" "left", Html.Attributes.style "padding" "12px" ] [ Html.text "External ID" ]
                , Html.th [ Html.Attributes.style "text-align" "left", Html.Attributes.style "padding" "12px" ] [ Html.text "Title" ]
                , Html.th [ Html.Attributes.style "text-align" "left", Html.Attributes.style "padding" "12px" ] [ Html.text "Status" ]
                , Html.th [ Html.Attributes.style "text-align" "left", Html.Attributes.style "padding" "12px" ] [ Html.text "Actions" ]
                ]
            ]
        , Html.tbody []
            (List.map viewIssueLinkRow links)
        ]


viewIssueLinkRow : IssueLink -> Html Msg
viewIssueLinkRow link =
    Html.tr [ Html.Attributes.style "border-bottom" "1px solid #ddd" ]
        [ Html.td [ Html.Attributes.style "padding" "12px" ]
            [ Html.span
                [ Html.Attributes.style "background"
                    (if link.provider == "jira" then
                        "#e3f2fd"

                     else
                        "#f3e5f5"
                    )
                , Html.Attributes.style "padding" "2px 8px"
                , Html.Attributes.style "border-radius" "4px"
                , Html.Attributes.style "font-size" "12px"
                , Html.Attributes.style "text-transform" "capitalize"
                ]
                [ Html.text link.provider ]
            ]
        , Html.td [ Html.Attributes.style "padding" "12px" ]
            [ if String.isEmpty link.url then
                Html.text link.externalId

              else
                Html.a
                    [ Html.Attributes.href link.url
                    , Html.Attributes.target "_blank"
                    , Html.Attributes.style "color" "#1976d2"
                    ]
                    [ Html.text link.externalId ]
            ]
        , Html.td [ Html.Attributes.style "padding" "12px" ] [ Html.text link.title ]
        , Html.td [ Html.Attributes.style "padding" "12px" ]
            [ Html.span
                [ Html.Attributes.style "font-weight" "bold"
                , Html.Attributes.style "font-size" "13px"
                ]
                [ Html.text link.status ]
            ]
        , Html.td [ Html.Attributes.style "padding" "12px" ]
            [ Html.button
                [ Html.Events.onClick (ResolveIssue link.id)
                , Html.Attributes.class "mdc-button"
                , Html.Attributes.style "color" "#388e3c"
                , Html.Attributes.style "font-size" "12px"
                ]
                [ Html.text "Resolve" ]
            , Html.button
                [ Html.Events.onClick (SyncIssue link.id)
                , Html.Attributes.class "mdc-button"
                , Html.Attributes.style "color" "#1976d2"
                , Html.Attributes.style "font-size" "12px"
                ]
                [ Html.text "Sync" ]
            , Html.button
                [ Html.Events.onClick (UnlinkIssue link.id)
                , Html.Attributes.class "mdc-button"
                , Html.Attributes.style "color" "#f44336"
                , Html.Attributes.style "font-size" "12px"
                ]
                [ Html.text "Unlink" ]
            ]
        ]


viewCreateIssueDialog : List Integration -> CreateIssueDialogState -> Html Msg
viewCreateIssueDialog integrations dialog =
    let
        selectedProvider =
            List.filter (\i -> i.id == dialog.integrationId) integrations
                |> List.head
                |> Maybe.map .provider
                |> Maybe.withDefault ""
    in
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
            , Html.Attributes.style "min-width" "400px"
            , Html.Attributes.style "max-width" "500px"
            ]
            [ Html.h2
                [ Html.Attributes.class "mdc-typography--headline6"
                , Html.Attributes.style "margin-top" "0"
                ]
                [ Html.text "Create Issue" ]
            , Html.div
                [ Html.Attributes.style "margin-bottom" "16px" ]
                [ Html.label
                    [ Html.Attributes.style "display" "block"
                    , Html.Attributes.style "margin-bottom" "4px"
                    ]
                    [ Html.text "Integration" ]
                , Html.select
                    [ Html.Events.onInput SetCreateIssueIntegration
                    , Html.Attributes.style "width" "100%"
                    , Html.Attributes.style "padding" "8px"
                    , Html.Attributes.style "border" "1px solid #ccc"
                    , Html.Attributes.style "border-radius" "4px"
                    ]
                    (List.map
                        (\integration ->
                            Html.option
                                [ Html.Attributes.value integration.id
                                , Html.Attributes.selected (integration.id == dialog.integrationId)
                                ]
                                [ Html.text (integration.name ++ " (" ++ integration.provider ++ ")") ]
                        )
                        integrations
                    )
                ]
            , Html.div
                [ Html.Attributes.style "margin-bottom" "16px" ]
                [ Html.label
                    [ Html.Attributes.style "display" "block"
                    , Html.Attributes.style "margin-bottom" "4px"
                    ]
                    [ Html.text "Title" ]
                , Html.input
                    [ Html.Attributes.type_ "text"
                    , Html.Attributes.value dialog.title
                    , Html.Events.onInput SetCreateIssueTitle
                    , Html.Attributes.placeholder "Issue title"
                    , Html.Attributes.style "width" "100%"
                    , Html.Attributes.style "padding" "8px"
                    , Html.Attributes.style "border" "1px solid #ccc"
                    , Html.Attributes.style "border-radius" "4px"
                    , Html.Attributes.style "box-sizing" "border-box"
                    ]
                    []
                ]
            , Html.div
                [ Html.Attributes.style "margin-bottom" "16px" ]
                [ Html.label
                    [ Html.Attributes.style "display" "block"
                    , Html.Attributes.style "margin-bottom" "4px"
                    ]
                    [ Html.text "Description" ]
                , Html.textarea
                    [ Html.Attributes.value dialog.description
                    , Html.Events.onInput SetCreateIssueDescription
                    , Html.Attributes.placeholder "Issue description"
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
            , if selectedProvider == "jira" then
                Html.div []
                    [ Html.div
                        [ Html.Attributes.style "margin-bottom" "16px" ]
                        [ Html.label
                            [ Html.Attributes.style "display" "block"
                            , Html.Attributes.style "margin-bottom" "4px"
                            ]
                            [ Html.text "Project Key" ]
                        , Html.input
                            [ Html.Attributes.type_ "text"
                            , Html.Attributes.value dialog.projectKey
                            , Html.Events.onInput SetCreateIssueProjectKey
                            , Html.Attributes.placeholder "e.g., PROJ"
                            , Html.Attributes.style "width" "100%"
                            , Html.Attributes.style "padding" "8px"
                            , Html.Attributes.style "border" "1px solid #ccc"
                            , Html.Attributes.style "border-radius" "4px"
                            , Html.Attributes.style "box-sizing" "border-box"
                            ]
                            []
                        ]
                    , Html.div
                        [ Html.Attributes.style "margin-bottom" "16px" ]
                        [ Html.label
                            [ Html.Attributes.style "display" "block"
                            , Html.Attributes.style "margin-bottom" "4px"
                            ]
                            [ Html.text "Issue Type" ]
                        , Html.select
                            [ Html.Events.onInput SetCreateIssueType
                            , Html.Attributes.style "width" "100%"
                            , Html.Attributes.style "padding" "8px"
                            , Html.Attributes.style "border" "1px solid #ccc"
                            , Html.Attributes.style "border-radius" "4px"
                            ]
                            [ Html.option [ Html.Attributes.value "Bug", Html.Attributes.selected (dialog.issueType == "Bug") ] [ Html.text "Bug" ]
                            , Html.option [ Html.Attributes.value "Task", Html.Attributes.selected (dialog.issueType == "Task") ] [ Html.text "Task" ]
                            , Html.option [ Html.Attributes.value "Story", Html.Attributes.selected (dialog.issueType == "Story") ] [ Html.text "Story" ]
                            ]
                        ]
                    ]

              else if selectedProvider == "github" then
                Html.div
                    [ Html.Attributes.style "margin-bottom" "16px" ]
                    [ Html.label
                        [ Html.Attributes.style "display" "block"
                        , Html.Attributes.style "margin-bottom" "4px"
                        ]
                        [ Html.text "Repository" ]
                    , Html.input
                        [ Html.Attributes.type_ "text"
                        , Html.Attributes.value dialog.repository
                        , Html.Events.onInput SetCreateIssueRepository
                        , Html.Attributes.placeholder "owner/repo"
                        , Html.Attributes.style "width" "100%"
                        , Html.Attributes.style "padding" "8px"
                        , Html.Attributes.style "border" "1px solid #ccc"
                        , Html.Attributes.style "border-radius" "4px"
                        , Html.Attributes.style "box-sizing" "border-box"
                        ]
                        []
                    ]

              else
                Html.text ""
            , Html.div
                [ Html.Attributes.style "display" "flex"
                , Html.Attributes.style "justify-content" "flex-end"
                , Html.Attributes.style "gap" "8px"
                ]
                [ Html.button
                    [ Html.Events.onClick CloseCreateIssueDialog
                    , Html.Attributes.class "mdc-button"
                    ]
                    [ Html.text "Cancel" ]
                , Html.button
                    [ Html.Events.onClick SubmitCreateIssue
                    , Html.Attributes.class "mdc-button mdc-button--raised"
                    ]
                    [ Html.text "Create" ]
                ]
            ]
        ]


viewLinkIssueDialog : List Integration -> LinkIssueDialogState -> Html Msg
viewLinkIssueDialog integrations dialog =
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
            , Html.Attributes.style "min-width" "400px"
            , Html.Attributes.style "max-width" "500px"
            ]
            [ Html.h2
                [ Html.Attributes.class "mdc-typography--headline6"
                , Html.Attributes.style "margin-top" "0"
                ]
                [ Html.text "Link Existing Issue" ]
            , Html.div
                [ Html.Attributes.style "margin-bottom" "16px" ]
                [ Html.label
                    [ Html.Attributes.style "display" "block"
                    , Html.Attributes.style "margin-bottom" "4px"
                    ]
                    [ Html.text "Integration" ]
                , Html.select
                    [ Html.Events.onInput SetLinkIssueIntegration
                    , Html.Attributes.style "width" "100%"
                    , Html.Attributes.style "padding" "8px"
                    , Html.Attributes.style "border" "1px solid #ccc"
                    , Html.Attributes.style "border-radius" "4px"
                    ]
                    (List.map
                        (\integration ->
                            Html.option
                                [ Html.Attributes.value integration.id
                                , Html.Attributes.selected (integration.id == dialog.integrationId)
                                ]
                                [ Html.text (integration.name ++ " (" ++ integration.provider ++ ")") ]
                        )
                        integrations
                    )
                ]
            , Html.div
                [ Html.Attributes.style "margin-bottom" "16px" ]
                [ Html.label
                    [ Html.Attributes.style "display" "block"
                    , Html.Attributes.style "margin-bottom" "4px"
                    ]
                    [ Html.text "Search Issues" ]
                , Html.div
                    [ Html.Attributes.style "display" "flex"
                    , Html.Attributes.style "gap" "8px"
                    ]
                    [ Html.input
                        [ Html.Attributes.type_ "text"
                        , Html.Attributes.placeholder "Search by title or ID..."
                        , Html.Attributes.value dialog.searchQuery
                        , Html.Events.onInput SetLinkIssueSearchQuery
                        , Html.Attributes.style "flex" "1"
                        , Html.Attributes.style "padding" "8px"
                        , Html.Attributes.style "border" "1px solid #ccc"
                        , Html.Attributes.style "border-radius" "4px"
                        , Html.Attributes.style "font-size" "14px"
                        ]
                        []
                    , Html.button
                        [ Html.Events.onClick SearchExternalIssues
                        , Html.Attributes.class "mdc-button mdc-button--outlined"
                        ]
                        [ Html.text "Search" ]
                    ]
                ]
            , if dialog.loading then
                Html.p
                    [ Html.Attributes.style "color" "#999"
                    , Html.Attributes.style "font-size" "14px"
                    , Html.Attributes.style "margin-bottom" "16px"
                    ]
                    [ Html.text "Searching..." ]

              else if not (List.isEmpty dialog.searchResults) then
                Html.div
                    [ Html.Attributes.style "max-height" "200px"
                    , Html.Attributes.style "overflow-y" "auto"
                    , Html.Attributes.style "border" "1px solid #ddd"
                    , Html.Attributes.style "border-radius" "4px"
                    , Html.Attributes.style "margin-bottom" "16px"
                    ]
                    (List.map (viewExternalIssueResult dialog.selectedIssue) dialog.searchResults)

              else
                Html.text ""
            , case dialog.selectedIssue of
                Just issue ->
                    Html.div
                        [ Html.Attributes.style "margin-bottom" "16px"
                        , Html.Attributes.style "padding" "8px"
                        , Html.Attributes.style "background" "#e8f5e9"
                        , Html.Attributes.style "border-radius" "4px"
                        ]
                        [ Html.text ("Selected: " ++ issue.externalId ++ " - " ++ issue.title) ]

                Nothing ->
                    Html.text ""
            , Html.div
                [ Html.Attributes.style "display" "flex"
                , Html.Attributes.style "justify-content" "flex-end"
                , Html.Attributes.style "gap" "8px"
                ]
                [ Html.button
                    [ Html.Events.onClick CloseLinkIssueDialog
                    , Html.Attributes.class "mdc-button"
                    ]
                    [ Html.text "Cancel" ]
                , Html.button
                    [ Html.Events.onClick SubmitLinkIssue
                    , Html.Attributes.class "mdc-button mdc-button--raised"
                    , Html.Attributes.disabled (dialog.selectedIssue == Nothing)
                    ]
                    [ Html.text "Link" ]
                ]
            ]
        ]


viewExternalIssueResult : Maybe ExternalIssue -> ExternalIssue -> Html Msg
viewExternalIssueResult selectedIssue issue =
    let
        isSelected =
            case selectedIssue of
                Just sel ->
                    sel.externalId == issue.externalId

                Nothing ->
                    False
    in
    Html.div
        [ Html.Events.onClick (SelectExternalIssue issue)
        , Html.Attributes.style "padding" "8px 12px"
        , Html.Attributes.style "cursor" "pointer"
        , Html.Attributes.style "border-bottom" "1px solid #eee"
        , Html.Attributes.style "background"
            (if isSelected then
                "#e3f2fd"

             else
                "white"
            )
        ]
        [ Html.div
            [ Html.Attributes.style "font-weight" "bold"
            , Html.Attributes.style "font-size" "14px"
            ]
            [ Html.text (issue.externalId ++ ": " ++ issue.title) ]
        , Html.div
            [ Html.Attributes.style "font-size" "12px"
            , Html.Attributes.style "color" "#666"
            ]
            [ Html.text ("Status: " ++ issue.status) ]
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
