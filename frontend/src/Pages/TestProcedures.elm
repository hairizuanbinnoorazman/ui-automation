module Pages.TestProcedures exposing (Model, Msg, init, update, view)

import API
import Dict exposing (Dict)
import File exposing (File)
import Components
import Html exposing (Html, button, div, h1, h2, h3, h4, input, label, li, p, span, text, textarea, ul)
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


type alias CreateDialogState =
    { name : String
    , description : String
    }


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
    , draftLoading : Bool
    , committedLoading : Bool
    , loading : Bool
    , error : Maybe String
    , createDialog : Maybe CreateDialogState
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
      , draftLoading = False
      , committedLoading = False
      , loading = False
      , error = Nothing
      , createDialog = Nothing
      }
    , API.getTestProcedures projectId 10 0 ProceduresResponse
    )



-- UPDATE


type Msg
    = ProceduresResponse (Result Http.Error (PaginatedResponse TestProcedure))
    | LoadPage Int
    | OpenCreateDialog
    | CloseCreateDialog
    | SetCreateName String
    | SetCreateDescription String
    | SubmitCreate
    | CreateResponse (Result Http.Error TestProcedure)
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

        OpenCreateDialog ->
            ( { model | createDialog = Just { name = "", description = "" } }
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

        SubmitCreate ->
            case model.createDialog of
                Just dialog ->
                    ( { model | loading = True, error = Nothing }
                    , API.createTestProcedure model.projectId
                        { name = dialog.name, description = dialog.description, steps = [] }
                        CreateResponse
                    )

                Nothing ->
                    ( model, Cmd.none )

        CreateResponse result ->
            case result of
                Ok _ ->
                    ( { model | createDialog = Nothing, loading = False }
                    , API.getTestProcedures model.projectId model.limit model.offset ProceduresResponse
                    )

                Err _ ->
                    ( { model | error = Just "Failed to create procedure", loading = False }
                    , Cmd.none
                    )

        SelectProcedure procedure ->
            ( { model
                | selectedProcedure = Just procedure
                , viewMode = ViewMode
                , draftProcedure = Nothing
                , committedProcedure = Nothing
                , draftLoading = True
                , committedLoading = True
                , error = Nothing
              }
            , Cmd.batch
                [ API.getTestProcedure model.projectId procedure.id True DraftResponse
                , API.getTestProcedure model.projectId procedure.id False CommittedResponse
                ]
            )

        SwitchToViewMode ->
            ( { model | viewMode = ViewMode, error = Nothing }, Cmd.none )

        SwitchToEditMode ->
            case model.draftProcedure of
                Just draft ->
                    ( { model
                        | viewMode = EditMode
                        , editingSteps = draft.steps
                        , error = Nothing
                      }
                    , Cmd.none
                    )

                Nothing ->
                    ( model, Cmd.none )

        SwitchToNewVersionMode ->
            case model.selectedProcedure of
                Just procedure ->
                    ( { model | viewMode = NewVersionMode, loading = True, error = Nothing }
                    , API.getDraftDiff procedure.id DiffResponse
                    )

                Nothing ->
                    ( model, Cmd.none )

        LoadDraftAndCommitted ->
            case model.selectedProcedure of
                Just procedure ->
                    ( { model | draftLoading = True, committedLoading = True }
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
                    ( { model | draftProcedure = Just draft, draftLoading = False }
                    , Cmd.none
                    )

                Err _ ->
                    ( { model | draftProcedure = Nothing, draftLoading = False }
                    , Cmd.none
                    )

        CommittedResponse result ->
            case result of
                Ok committed ->
                    ( { model | committedProcedure = Just committed, committedLoading = False }
                    , Cmd.none
                    )

                Err _ ->
                    ( { model | committedProcedure = Nothing, committedLoading = False }
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
                            ( { model | loading = True, error = Nothing }
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
                    ( { model | loading = True, error = Nothing }
                    , API.resetDraft procedure.id DraftReset
                    )

                Nothing ->
                    ( model, Cmd.none )

        DraftReset result ->
            case result of
                Ok () ->
                    update LoadDraftAndCommitted { model | error = Nothing }

                Err _ ->
                    ( { model | error = Just "Failed to reset draft", loading = False }
                    , Cmd.none
                    )

        CommitVersion ->
            case model.selectedProcedure of
                Just procedure ->
                    ( { model | loading = True, error = Nothing }
                    , API.commitDraft procedure.id VersionCommitted
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
                        , error = Nothing
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
    div []
        [ div [ class "page-header" ]
            [ h1 [ class "mdc-typography--headline3" ] [ text "Test Procedures" ]
            , button
                [ onClick OpenCreateDialog
                , class "mdc-button mdc-button--raised"
                ]
                [ text "New Procedure" ]
            ]
        , case model.error of
            Just err ->
                div
                    [ style "color" "red"
                    , style "margin-bottom" "20px"
                    ]
                    [ text err ]

            Nothing ->
                text ""
        , div [ class "procedures-layout" ]
            [ viewProcedureList model
            , viewSelectedProcedure model
            ]
        , case model.createDialog of
            Just dialog ->
                viewCreateDialog dialog

            Nothing ->
                text ""
        ]


viewCreateDialog : CreateDialogState -> Html Msg
viewCreateDialog dialog =
    Components.viewDialogOverlay "Create Procedure"
        [ Components.viewFormField "Name"
            [ type_ "text"
            , placeholder "Procedure name"
            , value dialog.name
            , onInput SetCreateName
            ]
        , Components.viewFormField "Description"
            [ type_ "text"
            , placeholder "Procedure description"
            , value dialog.description
            , onInput SetCreateDescription
            ]
        ]
        [ button
            [ onClick CloseCreateDialog
            , class "mdc-button"
            ]
            [ text "Cancel" ]
        , button
            [ onClick SubmitCreate
            , class "mdc-button mdc-button--raised"
            , disabled (String.isEmpty dialog.name)
            ]
            [ text "Create" ]
        ]


viewProcedureList : Model -> Html Msg
viewProcedureList model =
    div [ class "procedures-list" ]
        [ if List.isEmpty model.procedures then
            p [ class "mdc-typography--body1" ] [ text "No procedures found" ]

          else
            Html.table
                [ class "mdc-data-table__table"
                , style "width" "100%"
                , style "border-collapse" "collapse"
                ]
                [ Html.thead []
                    [ Html.tr []
                        [ Html.th [ style "text-align" "left", style "padding" "12px" ] [ text "Name" ]
                        ]
                    ]
                , Html.tbody []
                    (List.map
                        (\proc ->
                            Html.tr
                                [ onClick (SelectProcedure proc)
                                , style "border-bottom" "1px solid #ddd"
                                , style "cursor" "pointer"
                                ]
                                [ Html.td [ style "padding" "12px" ] [ text proc.name ]
                                ]
                        )
                        model.procedures
                    )
                ]
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
    div
        [ style "display" "flex"
        , style "justify-content" "center"
        , style "align-items" "center"
        , style "gap" "10px"
        , style "margin-top" "20px"
        ]
        [ button
            [ onClick (LoadPage (max 0 (model.offset - model.limit)))
            , disabled (currentPage == 0)
            , class "mdc-button"
            ]
            [ text "Previous" ]
        , span [] [ text ("Page " ++ String.fromInt (currentPage + 1) ++ " of " ++ String.fromInt (max 1 totalPages)) ]
        , button
            [ onClick (LoadPage (model.offset + model.limit))
            , disabled (currentPage >= totalPages - 1)
            , class "mdc-button"
            ]
            [ text "Next" ]
        ]


viewSelectedProcedure : Model -> Html Msg
viewSelectedProcedure model =
    case model.selectedProcedure of
        Nothing ->
            div
                [ style "padding" "24px" ]
                [ p [ class "mdc-typography--body1" ] [ text "Select a procedure to view details" ]
                ]

        Just _ ->
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


procedureContentEqual : Maybe TestProcedure -> Maybe TestProcedure -> Bool
procedureContentEqual maybeA maybeB =
    case ( maybeA, maybeB ) of
        ( Nothing, Nothing ) ->
            True

        ( Just a, Just b ) ->
            a.name == b.name && a.description == b.description && a.steps == b.steps

        _ ->
            False


viewModeSelector : Model -> Html Msg
viewModeSelector model =
    div
        [ style "display" "flex"
        , style "gap" "8px"
        , style "margin-bottom" "16px"
        ]
        [ button
            [ onClick SwitchToViewMode
            , class
                (if model.viewMode == ViewMode then
                    "mdc-button mdc-button--raised"

                 else
                    "mdc-button"
                )
            ]
            [ text "View" ]
        , button
            [ onClick SwitchToEditMode
            , class
                (if model.viewMode == EditMode then
                    "mdc-button mdc-button--raised"

                 else
                    "mdc-button"
                )
            ]
            [ text "Edit" ]
        , button
            [ onClick SwitchToNewVersionMode
            , class
                (if model.viewMode == NewVersionMode then
                    "mdc-button mdc-button--raised"

                 else
                    "mdc-button"
                )
            , disabled (procedureContentEqual model.draftProcedure model.committedProcedure)
            ]
            [ text "New Version" ]
        ]


viewModeView : Model -> Html Msg
viewModeView model =
    if model.draftLoading || model.committedLoading then
        div [] [ text "Loading..." ]

    else
        case ( model.committedProcedure, model.draftProcedure ) of
            ( Just committed, _ ) ->
                div []
                    [ div
                        [ style "display" "flex"
                        , style "align-items" "center"
                        , style "gap" "8px"
                        , style "margin-bottom" "4px"
                        ]
                        [ h3
                            [ class "mdc-typography--headline5"
                            , style "margin" "0"
                            ]
                            [ text committed.name ]
                        , span
                            [ style "background-color" "#1976d2"
                            , style "color" "white"
                            , style "font-size" "12px"
                            , style "font-weight" "500"
                            , style "padding" "2px 8px"
                            , style "border-radius" "12px"
                            ]
                            [ text ("v" ++ String.fromInt committed.version) ]
                        ]
                    , p [ class "mdc-typography--body1" ] [ text committed.description ]
                    , viewSteps committed.steps
                    ]

            ( Nothing, Just draft ) ->
                div []
                    [ div
                        [ style "background-color" "#fff3e0"
                        , style "border-left" "4px solid #ff9800"
                        , style "padding" "12px 16px"
                        , style "margin-bottom" "16px"
                        , class "mdc-typography--body2"
                        ]
                        [ text "Draft only - No published version yet" ]
                    , h3 [ class "mdc-typography--headline5" ] [ text draft.name ]
                    , p [ class "mdc-typography--body1" ] [ text draft.description ]
                    , viewSteps draft.steps
                    ]

            _ ->
                div [] [ text "No data available" ]


viewSteps : List TestStep -> Html Msg
viewSteps steps =
    if List.isEmpty steps then
        p [ class "mdc-typography--body1" ] [ text "No steps defined" ]

    else
        div []
            (List.indexedMap
                (\index step ->
                    div
                        [ style "border" "1px solid #e0e0e0"
                        , style "border-radius" "4px"
                        , style "padding" "16px"
                        , style "margin-bottom" "12px"
                        ]
                        [ h4 [ class "mdc-typography--subtitle1", style "margin-top" "0" ]
                            [ text (String.fromInt (index + 1) ++ ". " ++ step.name) ]
                        , p [ class "mdc-typography--body2" ] [ text step.instructions ]
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
        div
            [ style "display" "flex"
            , style "flex-wrap" "wrap"
            , style "gap" "8px"
            , style "margin-top" "8px"
            ]
            (List.map
                (\path ->
                    Html.img
                        [ Html.Attributes.src ("/uploads/" ++ path)
                        , style "max-width" "120px"
                        , style "border-radius" "4px"
                        , style "border" "1px solid #e0e0e0"
                        ]
                        []
                )
                imagePaths
            )


viewEditMode : Model -> Html Msg
viewEditMode model =
    case model.draftProcedure of
        Nothing ->
            div [] [ text "Loading draft..." ]

        Just draft ->
            div []
                [ h3 [ class "mdc-typography--headline5" ] [ text "Edit Draft" ]
                , div [ style "margin-bottom" "16px" ]
                    [ Components.viewFormField "Name"
                        [ type_ "text"
                        , value draft.name
                        , disabled True
                        ]
                    , Components.viewFormField "Description"
                        [ type_ "text"
                        , value draft.description
                        , disabled True
                        ]
                    ]
                , div []
                    (List.indexedMap (viewEditableStep model) model.editingSteps)
                , button
                    [ onClick AddStep
                    , class "mdc-button"
                    , style "margin-bottom" "16px"
                    ]
                    [ text "+ Add Step" ]
                , div
                    [ style "display" "flex"
                    , style "gap" "8px"
                    ]
                    [ button [ onClick SaveDraft, class "mdc-button mdc-button--raised" ] [ text "Save Draft" ]
                    , button [ onClick ClearChanges, class "mdc-button" ] [ text "Clear Changes" ]
                    , button [ onClick SwitchToViewMode, class "mdc-button" ] [ text "Done Editing" ]
                    ]
                ]


viewEditableStep : Model -> Int -> TestStep -> Html Msg
viewEditableStep model index step =
    div
        [ style "border" "1px solid #e0e0e0"
        , style "border-radius" "4px"
        , style "padding" "16px"
        , style "margin-bottom" "12px"
        ]
        [ Components.viewFormField "Step Name"
            [ type_ "text"
            , placeholder "Step name"
            , value step.name
            , onInput (UpdateStepName index)
            ]
        , Components.viewTextArea "Instructions"
            [ placeholder "Instructions"
            , value step.instructions
            , onInput (UpdateStepInstructions index)
            ]
        , div [ style "margin-bottom" "12px" ]
            [ input
                [ type_ "file"
                , Html.Attributes.accept "image/*"
                , on "change" (Decode.map (ImageSelected index) fileDecoder)
                ]
                []
            , if Dict.member index model.uploadingImages then
                span [ class "mdc-typography--caption", style "margin-left" "8px" ] [ text "Uploading..." ]

              else
                text ""
            ]
        , div
            [ style "display" "flex"
            , style "flex-wrap" "wrap"
            , style "gap" "8px"
            , style "margin-bottom" "12px"
            ]
            (List.indexedMap
                (\imgIdx path ->
                    div []
                        [ Html.img
                            [ Html.Attributes.src ("/uploads/" ++ path)
                            , style "max-width" "100px"
                            , style "border-radius" "4px"
                            ]
                            []
                        , button
                            [ onClick (RemoveStepImage index imgIdx)
                            , class "mdc-button"
                            , style "display" "block"
                            ]
                            [ text "Remove" ]
                        ]
                )
                step.imagePaths
            )
        , button
            [ onClick (RemoveStep index)
            , class "mdc-button"
            , style "color" "#d32f2f"
            ]
            [ text "Delete Step" ]
        ]


fileDecoder : Decode.Decoder File
fileDecoder =
    Decode.at [ "target", "files", "0" ] File.decoder


viewNewVersionMode : Model -> Html Msg
viewNewVersionMode model =
    let
        nextVersionNumber =
            case model.committedProcedure of
                Just committed ->
                    committed.version + 1

                Nothing ->
                    1

        currentVersionLabel =
            case model.committedProcedure of
                Just committed ->
                    "Current Version (v" ++ String.fromInt committed.version ++ ")"

                Nothing ->
                    "Current Version (none)"

        draftVersionLabel =
            "Draft Changes -> will become v" ++ String.fromInt nextVersionNumber
    in
    div []
        [ h3 [ class "mdc-typography--headline5" ]
            [ text ("Creating Version " ++ String.fromInt nextVersionNumber) ]
        , div
            [ style "display" "flex"
            , style "gap" "16px"
            ]
            [ div [ style "flex" "1" ]
                [ h4 [ class "mdc-typography--subtitle1" ] [ text currentVersionLabel ]
                , case model.committedProcedure of
                    Just committed ->
                        div []
                            [ p
                                [ class "mdc-typography--body2"
                                , style "color" "#555"
                                , style "margin-bottom" "12px"
                                ]
                                [ text committed.description ]
                            , viewSteps committed.steps
                            ]

                    Nothing ->
                        p [ class "mdc-typography--body1" ] [ text "No published version" ]
                ]
            , div [ style "flex" "1" ]
                [ h4 [ class "mdc-typography--subtitle1" ] [ text draftVersionLabel ]
                , case model.draftProcedure of
                    Just draft ->
                        div []
                            [ p
                                [ class "mdc-typography--body2"
                                , style "color" "#555"
                                , style "margin-bottom" "12px"
                                ]
                                [ text draft.description ]
                            , viewSteps draft.steps
                            ]

                    Nothing ->
                        p [ class "mdc-typography--body1" ] [ text "No draft" ]
                ]
            ]
        , div
            [ style "display" "flex"
            , style "gap" "8px"
            , style "margin-top" "16px"
            ]
            [ button [ onClick SwitchToViewMode, class "mdc-button" ] [ text "Cancel" ]
            , button [ onClick CommitVersion, class "mdc-button mdc-button--raised" ] [ text "Create New Version" ]
            ]
        ]


viewError : Maybe String -> Html Msg
viewError maybeError =
    case maybeError of
        Nothing ->
            text ""

        Just errorMsg ->
            div
                [ style "color" "red"
                , style "margin-top" "16px"
                ]
                [ text errorMsg ]
