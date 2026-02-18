module Pages.TestProcedures exposing (Model, Msg, init, update, view)

import API
import Dict exposing (Dict)
import File exposing (File)
import Html exposing (Html, button, div, h2, h3, h4, input, li, p, span, text, textarea, ul)
import Html.Attributes exposing (class, disabled, placeholder, style, type_, value)
import Html.Events exposing (on, onClick, onInput)
import Http
import Json.Decode as Decode
import Types exposing (DraftDiff, PaginatedResponse, TestProcedure, TestStep)


-- MODEL


type ProcedureViewMode
    = ViewMode
    | EditMode
    | NewVersionMode


type alias Model =
    { projectId : String
    , procedures : List TestProcedure
    , total : Int
    , limit : Int
    , offset : Int
    , selectedProcedure : Maybe TestProcedure
    , viewMode : ProcedureViewMode
    , draftProcedure : Maybe TestProcedure
    , committedProcedure : Maybe TestProcedure
    , editingSteps : List TestStep
    , uploadingImages : Dict Int Bool
    , loading : Bool
    , error : Maybe String
    }


init : String -> ( Model, Cmd Msg )
init projectId =
    ( { projectId = projectId
      , procedures = []
      , total = 0
      , limit = 10
      , offset = 0
      , selectedProcedure = Nothing
      , viewMode = ViewMode
      , draftProcedure = Nothing
      , committedProcedure = Nothing
      , editingSteps = []
      , uploadingImages = Dict.empty
      , loading = False
      , error = Nothing
      }
    , API.getTestProcedures projectId 10 0 ProceduresResponse
    )



-- UPDATE


type Msg
    = ProceduresResponse (Result Http.Error (PaginatedResponse TestProcedure))
    | LoadPage Int
    | SelectProcedure TestProcedure
    | SwitchToViewMode
    | SwitchToEditMode
    | SwitchToNewVersionMode
    | LoadDraftAndCommitted
    | DraftResponse (Result Http.Error TestProcedure)
    | CommittedResponse (Result Http.Error TestProcedure)
    | DiffResponse (Result Http.Error DraftDiff)
    | AddStep
    | RemoveStep Int
    | UpdateStepName Int String
    | UpdateStepInstructions Int String
    | ImageSelected Int File
    | ImageUploaded Int (Result Http.Error String)
    | RemoveStepImage Int Int
    | SaveDraft
    | DraftSaved (Result Http.Error TestProcedure)
    | ClearChanges
    | DraftReset (Result Http.Error ())
    | CommitVersion
    | VersionCommitted (Result Http.Error TestProcedure)


update : Msg -> Model -> ( Model, Cmd Msg )
update msg model =
    case msg of
        ProceduresResponse result ->
            case result of
                Ok response ->
                    ( { model
                        | procedures = response.items
                        , total = response.total
                        , loading = False
                      }
                    , Cmd.none
                    )

                Err _ ->
                    ( { model | error = Just "Failed to load procedures", loading = False }
                    , Cmd.none
                    )

        LoadPage offset ->
            ( { model | offset = offset, loading = True }
            , API.getTestProcedures model.projectId model.limit offset ProceduresResponse
            )

        SelectProcedure procedure ->
            ( { model
                | selectedProcedure = Just procedure
                , viewMode = ViewMode
                , loading = True
              }
            , Cmd.batch
                [ API.getTestProcedure model.projectId procedure.id True DraftResponse
                , API.getTestProcedure model.projectId procedure.id False CommittedResponse
                ]
            )

        SwitchToViewMode ->
            ( { model | viewMode = ViewMode }, Cmd.none )

        SwitchToEditMode ->
            case model.draftProcedure of
                Just draft ->
                    ( { model
                        | viewMode = EditMode
                        , editingSteps = draft.steps
                      }
                    , Cmd.none
                    )

                Nothing ->
                    ( model, Cmd.none )

        SwitchToNewVersionMode ->
            case model.selectedProcedure of
                Just procedure ->
                    ( { model | viewMode = NewVersionMode, loading = True }
                    , API.getDraftDiff model.projectId procedure.id DiffResponse
                    )

                Nothing ->
                    ( model, Cmd.none )

        LoadDraftAndCommitted ->
            case model.selectedProcedure of
                Just procedure ->
                    ( { model | loading = True }
                    , Cmd.batch
                        [ API.getTestProcedure model.projectId procedure.id True DraftResponse
                        , API.getTestProcedure model.projectId procedure.id False CommittedResponse
                        ]
                    )

                Nothing ->
                    ( model, Cmd.none )

        DraftResponse result ->
            case result of
                Ok draft ->
                    ( { model | draftProcedure = Just draft, loading = False }
                    , Cmd.none
                    )

                Err _ ->
                    ( { model | draftProcedure = Nothing, loading = False }
                    , Cmd.none
                    )

        CommittedResponse result ->
            case result of
                Ok committed ->
                    ( { model | committedProcedure = Just committed, loading = False }
                    , Cmd.none
                    )

                Err _ ->
                    ( { model | committedProcedure = Nothing, loading = False }
                    , Cmd.none
                    )

        DiffResponse result ->
            case result of
                Ok diff ->
                    ( { model
                        | draftProcedure = diff.draft
                        , committedProcedure = diff.committed
                        , loading = False
                      }
                    , Cmd.none
                    )

                Err _ ->
                    ( { model | error = Just "Failed to load diff", loading = False }
                    , Cmd.none
                    )

        AddStep ->
            ( { model
                | editingSteps =
                    model.editingSteps
                        ++ [ { name = "", instructions = "", imagePaths = [] } ]
              }
            , Cmd.none
            )

        RemoveStep index ->
            ( { model
                | editingSteps =
                    List.take index model.editingSteps
                        ++ List.drop (index + 1) model.editingSteps
              }
            , Cmd.none
            )

        UpdateStepName index newName ->
            ( { model
                | editingSteps =
                    List.indexedMap
                        (\i step ->
                            if i == index then
                                { step | name = newName }

                            else
                                step
                        )
                        model.editingSteps
              }
            , Cmd.none
            )

        UpdateStepInstructions index newInstructions ->
            ( { model
                | editingSteps =
                    List.indexedMap
                        (\i step ->
                            if i == index then
                                { step | instructions = newInstructions }

                            else
                                step
                        )
                        model.editingSteps
              }
            , Cmd.none
            )

        ImageSelected index file ->
            case model.selectedProcedure of
                Just procedure ->
                    ( { model | uploadingImages = Dict.insert index True model.uploadingImages }
                    , API.uploadStepImage procedure.id file (ImageUploaded index)
                    )

                Nothing ->
                    ( model, Cmd.none )

        ImageUploaded index result ->
            case result of
                Ok imagePath ->
                    ( { model
                        | editingSteps =
                            List.indexedMap
                                (\i step ->
                                    if i == index then
                                        { step | imagePaths = step.imagePaths ++ [ imagePath ] }

                                    else
                                        step
                                )
                                model.editingSteps
                        , uploadingImages = Dict.remove index model.uploadingImages
                      }
                    , Cmd.none
                    )

                Err _ ->
                    ( { model
                        | error = Just "Failed to upload image"
                        , uploadingImages = Dict.remove index model.uploadingImages
                      }
                    , Cmd.none
                    )

        RemoveStepImage stepIndex imageIndex ->
            ( { model
                | editingSteps =
                    List.indexedMap
                        (\i step ->
                            if i == stepIndex then
                                { step
                                    | imagePaths =
                                        List.take imageIndex step.imagePaths
                                            ++ List.drop (imageIndex + 1) step.imagePaths
                                }

                            else
                                step
                        )
                        model.editingSteps
              }
            , Cmd.none
            )

        SaveDraft ->
            case model.selectedProcedure of
                Just procedure ->
                    case model.draftProcedure of
                        Just draft ->
                            let
                                input =
                                    { name = draft.name
                                    , description = draft.description
                                    , steps = model.editingSteps
                                    }
                            in
                            ( { model | loading = True }
                            , API.updateTestProcedure model.projectId procedure.id input DraftSaved
                            )

                        Nothing ->
                            ( model, Cmd.none )

                Nothing ->
                    ( model, Cmd.none )

        DraftSaved result ->
            case result of
                Ok draft ->
                    ( { model
                        | draftProcedure = Just draft
                        , loading = False
                        , error = Nothing
                      }
                    , Cmd.none
                    )

                Err _ ->
                    ( { model | error = Just "Failed to save draft", loading = False }
                    , Cmd.none
                    )

        ClearChanges ->
            case model.selectedProcedure of
                Just procedure ->
                    ( { model | loading = True }
                    , API.resetDraft model.projectId procedure.id DraftReset
                    )

                Nothing ->
                    ( model, Cmd.none )

        DraftReset result ->
            case result of
                Ok () ->
                    update LoadDraftAndCommitted model

                Err _ ->
                    ( { model | error = Just "Failed to reset draft", loading = False }
                    , Cmd.none
                    )

        CommitVersion ->
            case model.selectedProcedure of
                Just procedure ->
                    ( { model | loading = True }
                    , API.commitDraft model.projectId procedure.id VersionCommitted
                    )

                Nothing ->
                    ( model, Cmd.none )

        VersionCommitted result ->
            case result of
                Ok newVersion ->
                    ( { model
                        | committedProcedure = Just newVersion
                        , viewMode = ViewMode
                        , loading = False
                      }
                    , API.getTestProcedures model.projectId model.limit model.offset ProceduresResponse
                    )

                Err _ ->
                    ( { model | error = Just "Failed to commit version", loading = False }
                    , Cmd.none
                    )



-- VIEW


view : Model -> Html Msg
view model =
    div [ class "test-procedures-page" ]
        [ h2 [] [ text "Test Procedures" ]
        , div [ class "procedures-layout" ]
            [ viewProcedureList model
            , viewSelectedProcedure model
            ]
        ]


viewProcedureList : Model -> Html Msg
viewProcedureList model =
    div [ class "procedures-list" ]
        [ h3 [] [ text "Procedures" ]
        , if List.isEmpty model.procedures then
            p [] [ text "No procedures found" ]

          else
            ul []
                (List.map
                    (\proc ->
                        li
                            [ onClick (SelectProcedure proc)
                            , class "procedure-item"
                            ]
                            [ text proc.name ]
                    )
                    model.procedures
                )
        , viewPagination model
        ]


viewPagination : Model -> Html Msg
viewPagination model =
    let
        currentPage =
            model.offset // model.limit

        totalPages =
            (model.total + model.limit - 1) // model.limit
    in
    div [ class "pagination" ]
        [ button
            [ onClick (LoadPage (model.offset - model.limit))
            , disabled (currentPage == 0)
            ]
            [ text "Previous" ]
        , span [] [ text ("Page " ++ String.fromInt (currentPage + 1) ++ " of " ++ String.fromInt totalPages) ]
        , button
            [ onClick (LoadPage (model.offset + model.limit))
            , disabled (currentPage >= totalPages - 1)
            ]
            [ text "Next" ]
        ]


viewSelectedProcedure : Model -> Html Msg
viewSelectedProcedure model =
    case model.selectedProcedure of
        Nothing ->
            div [ class "no-selection" ]
                [ p [] [ text "Select a procedure to view details" ]
                ]

        Just procedure ->
            div [ class "procedure-details" ]
                [ viewModeSelector model
                , case model.viewMode of
                    ViewMode ->
                        viewModeView model

                    EditMode ->
                        viewEditMode model

                    NewVersionMode ->
                        viewNewVersionMode model
                , viewError model.error
                ]


viewModeSelector : Model -> Html Msg
viewModeSelector model =
    div [ class "mode-selector" ]
        [ button
            [ onClick SwitchToViewMode
            , class
                (if model.viewMode == ViewMode then
                    "active"

                 else
                    ""
                )
            ]
            [ text "View" ]
        , button
            [ onClick SwitchToEditMode
            , class
                (if model.viewMode == EditMode then
                    "active"

                 else
                    ""
                )
            ]
            [ text "Edit" ]
        , button
            [ onClick SwitchToNewVersionMode
            , class
                (if model.viewMode == NewVersionMode then
                    "active"

                 else
                    ""
                )
            , disabled (model.draftProcedure == model.committedProcedure)
            ]
            [ text "New Version" ]
        ]


viewModeView : Model -> Html Msg
viewModeView model =
    case ( model.committedProcedure, model.draftProcedure ) of
        ( Just committed, _ ) ->
            div [ class "view-mode" ]
                [ h3 [] [ text committed.name ]
                , p [] [ text committed.description ]
                , viewSteps committed.steps
                ]

        ( Nothing, Just draft ) ->
            div [ class "view-mode draft-only" ]
                [ div [ class "draft-banner" ]
                    [ text "⚠ Draft only - No published version yet" ]
                , h3 [] [ text draft.name ]
                , p [] [ text draft.description ]
                , viewSteps draft.steps
                ]

        _ ->
            div [] [ text "Loading..." ]


viewSteps : List TestStep -> Html Msg
viewSteps steps =
    if List.isEmpty steps then
        p [ class "no-steps" ] [ text "No steps defined" ]

    else
        div [ class "steps-list" ]
            (List.indexedMap
                (\index step ->
                    div [ class "step-card" ]
                        [ h4 [] [ text (String.fromInt (index + 1) ++ ". " ++ step.name) ]
                        , p [] [ text step.instructions ]
                        , viewImageGallery step.imagePaths
                        ]
                )
                steps
            )


viewImageGallery : List String -> Html Msg
viewImageGallery imagePaths =
    if List.isEmpty imagePaths then
        text ""

    else
        div [ class "image-gallery" ]
            (List.map
                (\path ->
                    div [ class "image-item" ]
                        [ Html.img [ Html.Attributes.src ("/uploads/" ++ path) ] []
                        ]
                )
                imagePaths
            )


viewEditMode : Model -> Html Msg
viewEditMode model =
    case model.draftProcedure of
        Nothing ->
            div [] [ text "Loading draft..." ]

        Just draft ->
            div [ class "edit-mode" ]
                [ h3 [] [ text "Edit Draft" ]
                , div [ class "editing-steps" ]
                    (List.indexedMap (viewEditableStep model) model.editingSteps)
                , button [ onClick AddStep, class "add-step-btn" ] [ text "+ Add Step" ]
                , div [ class "edit-actions" ]
                    [ button [ onClick SaveDraft ] [ text "Save Draft" ]
                    , button [ onClick ClearChanges ] [ text "Clear Changes" ]
                    , button [ onClick SwitchToViewMode ] [ text "Done Editing" ]
                    ]
                ]


viewEditableStep : Model -> Int -> TestStep -> Html Msg
viewEditableStep model index step =
    div [ class "editable-step" ]
        [ input
            [ type_ "text"
            , placeholder "Step name"
            , value step.name
            , onInput (UpdateStepName index)
            , class "step-name-input"
            ]
            []
        , textarea
            [ placeholder "Instructions"
            , value step.instructions
            , onInput (UpdateStepInstructions index)
            , class "step-instructions-input"
            ]
            []
        , div [ class "image-upload-zone" ]
            [ input
                [ type_ "file"
                , Html.Attributes.accept "image/*"
                , on "change" (Decode.map (ImageSelected index) fileDecoder)
                ]
                []
            , if Dict.member index model.uploadingImages then
                span [ class "uploading" ] [ text "Uploading..." ]

              else
                text ""
            ]
        , div [ class "step-images" ]
            (List.indexedMap
                (\imgIdx path ->
                    div [ class "step-image-item" ]
                        [ Html.img [ Html.Attributes.src ("/uploads/" ++ path), style "max-width" "100px" ] []
                        , button [ onClick (RemoveStepImage index imgIdx), class "remove-image-btn" ] [ text "×" ]
                        ]
                )
                step.imagePaths
            )
        , button [ onClick (RemoveStep index), class "remove-step-btn" ] [ text "Delete Step" ]
        ]


fileDecoder : Decode.Decoder File
fileDecoder =
    Decode.at [ "target", "files", "0" ] File.decoder


viewNewVersionMode : Model -> Html Msg
viewNewVersionMode model =
    div [ class "new-version-mode" ]
        [ h3 [] [ text "Review Changes" ]
        , div [ class "diff-view" ]
            [ div [ class "diff-column" ]
                [ h4 [] [ text "Current Version" ]
                , case model.committedProcedure of
                    Just committed ->
                        viewSteps committed.steps

                    Nothing ->
                        p [] [ text "No published version" ]
                ]
            , div [ class "diff-column" ]
                [ h4 [] [ text "Draft Changes" ]
                , case model.draftProcedure of
                    Just draft ->
                        viewSteps draft.steps

                    Nothing ->
                        p [] [ text "No draft" ]
                ]
            ]
        , div [ class "version-actions" ]
            [ button [ onClick SwitchToViewMode ] [ text "Cancel" ]
            , button [ onClick CommitVersion, class "commit-btn" ] [ text "Create New Version" ]
            ]
        ]


viewError : Maybe String -> Html Msg
viewError maybeError =
    case maybeError of
        Nothing ->
            text ""

        Just errorMsg ->
            div [ class "error-message" ]
                [ text errorMsg
                ]
