module Pages.ProcedureDetail exposing (Model, Msg, init, update, view)

import API
import Components
import Dict exposing (Dict)
import File exposing (File)
import Html exposing (Html, button, div, h3, h4, input, p, span, text, textarea)
import Html.Attributes exposing (class, disabled, placeholder, style, type_, value)
import Html.Events exposing (on, onClick, onInput)
import Http
import Json.Decode as Decode
import Types exposing (DraftDiff, TestProcedure, TestStep)


type StepChange
    = Added TestStep
    | Removed TestStep
    | Unchanged TestStep


-- MODEL


type ProcedureViewMode
    = ViewMode
    | EditMode
    | NewVersionMode


type alias Model =
    { projectId : String
    , procedureId : String
    , viewMode : ProcedureViewMode
    , draftProcedure : Maybe TestProcedure
    , committedProcedure : Maybe TestProcedure
    , editingSteps : List TestStep
    , uploadingImages : Dict Int Bool
    , draftLoading : Bool
    , committedLoading : Bool
    , loading : Bool
    , error : Maybe String
    }


init : String -> String -> ( Model, Cmd Msg )
init projectId procedureId =
    ( { projectId = projectId
      , procedureId = procedureId
      , viewMode = ViewMode
      , draftProcedure = Nothing
      , committedProcedure = Nothing
      , editingSteps = []
      , uploadingImages = Dict.empty
      , draftLoading = True
      , committedLoading = True
      , loading = False
      , error = Nothing
      }
    , Cmd.batch
        [ API.getTestProcedure projectId procedureId True DraftResponse
        , API.getTestProcedure projectId procedureId False CommittedResponse
        ]
    )



-- UPDATE


type Msg
    = SwitchToViewMode
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
            ( { model | viewMode = NewVersionMode, loading = True, error = Nothing }
            , API.getDraftDiff model.procedureId DiffResponse
            )

        LoadDraftAndCommitted ->
            ( { model | draftLoading = True, committedLoading = True }
            , Cmd.batch
                [ API.getTestProcedure model.projectId model.procedureId True DraftResponse
                , API.getTestProcedure model.projectId model.procedureId False CommittedResponse
                ]
            )

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
            ( { model | uploadingImages = Dict.insert index True model.uploadingImages }
            , API.uploadStepImage model.procedureId file (ImageUploaded index)
            )

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
                    , API.updateTestProcedure model.projectId model.procedureId input DraftSaved
                    )

                Nothing ->
                    ( model, Cmd.none )

        DraftSaved result ->
            case result of
                Ok draft ->
                    ( { model
                        | draftProcedure = Just draft
                        , editingSteps = draft.steps
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
            ( { model | loading = True, error = Nothing }
            , API.resetDraft model.procedureId DraftReset
            )

        DraftReset result ->
            case result of
                Ok () ->
                    update LoadDraftAndCommitted { model | error = Nothing }

                Err _ ->
                    ( { model | error = Just "Failed to reset draft", loading = False }
                    , Cmd.none
                    )

        CommitVersion ->
            ( { model | loading = True, error = Nothing }
            , API.commitDraft model.procedureId VersionCommitted
            )

        VersionCommitted result ->
            case result of
                Ok newVersion ->
                    ( { model
                        | committedProcedure = Just newVersion
                        , viewMode = ViewMode
                        , loading = False
                        , error = Nothing
                      }
                    , Cmd.none
                    )

                Err _ ->
                    ( { model | error = Just "Failed to commit version", loading = False }
                    , Cmd.none
                    )



-- VIEW


view : Model -> Html Msg
view model =
    div []
        [ Html.a
            [ Html.Attributes.href ("/projects/" ++ model.projectId ++ "/procedures")
            , style "display" "inline-block"
            , style "margin-bottom" "16px"
            , style "color" "#6200ee"
            , style "text-decoration" "none"
            ]
            [ text "← Back to Procedures" ]
        , div [ class "procedure-details" ]
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


computeStepDiff : List TestStep -> List TestStep -> List StepChange
computeStepDiff committedSteps draftSteps =
    let
        maxLen =
            max (List.length committedSteps) (List.length draftSteps)

        getAt idx lst =
            List.head (List.drop idx lst)

        buildDiff idx acc =
            if idx >= maxLen then
                List.reverse acc

            else
                let
                    changes =
                        case ( getAt idx committedSteps, getAt idx draftSteps ) of
                            ( Just c, Just d ) ->
                                if c.name == d.name && c.instructions == d.instructions then
                                    [ Unchanged d ]

                                else
                                    [ Removed c, Added d ]

                            ( Just c, Nothing ) ->
                                [ Removed c ]

                            ( Nothing, Just d ) ->
                                [ Added d ]

                            ( Nothing, Nothing ) ->
                                []
                in
                buildDiff (idx + 1) (List.reverse changes ++ acc)
    in
    buildDiff 0 []


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
            ( Just committed, Just draft ) ->
                let
                    hasDiff =
                        not (procedureContentEqual (Just committed) (Just draft))
                in
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
                    , if hasDiff then
                        div []
                            [ div
                                [ style "background-color" "#fff8e1"
                                , style "border-left" "4px solid #ff9800"
                                , style "padding" "12px 16px"
                                , style "margin-bottom" "16px"
                                , class "mdc-typography--body2"
                                ]
                                [ text "Showing draft with unpublished changes. Use 'New Version' to publish." ]
                            , viewStepsWithDiff (computeStepDiff committed.steps draft.steps)
                            ]

                      else
                        viewSteps committed.steps
                    ]

            ( Just committed, Nothing ) ->
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
                    , viewStepsWithDiff (List.map Added draft.steps)
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


viewStepsWithDiff : List StepChange -> Html Msg
viewStepsWithDiff changes =
    if List.isEmpty changes then
        p [ class "mdc-typography--body1" ] [ text "No steps defined" ]

    else
        let
            renderChange stepNum change =
                case change of
                    Unchanged step ->
                        ( stepNum + 1
                        , div
                            [ style "border" "1px solid #e0e0e0"
                            , style "border-radius" "4px"
                            , style "padding" "16px"
                            , style "margin-bottom" "12px"
                            ]
                            [ h4 [ class "mdc-typography--subtitle1", style "margin-top" "0" ]
                                [ text (String.fromInt stepNum ++ ". " ++ step.name) ]
                            , p [ class "mdc-typography--body2" ] [ text step.instructions ]
                            , viewImageGallery step.imagePaths
                            ]
                        )

                    Added step ->
                        ( stepNum + 1
                        , div
                            [ style "border" "1px solid #4caf50"
                            , style "background-color" "#f1f8e9"
                            , style "border-radius" "4px"
                            , style "padding" "16px"
                            , style "margin-bottom" "12px"
                            ]
                            [ h4 [ class "mdc-typography--subtitle1", style "margin-top" "0" ]
                                [ span
                                    [ style "background-color" "#4caf50"
                                    , style "color" "white"
                                    , style "font-size" "11px"
                                    , style "font-weight" "bold"
                                    , style "padding" "1px 6px"
                                    , style "border-radius" "10px"
                                    , style "margin-right" "6px"
                                    ]
                                    [ text "+" ]
                                , text (String.fromInt stepNum ++ ". " ++ step.name)
                                ]
                            , p [ class "mdc-typography--body2" ] [ text step.instructions ]
                            , viewImageGallery step.imagePaths
                            ]
                        )

                    Removed step ->
                        ( stepNum
                        , div
                            [ style "border" "1px solid #ef9a9a"
                            , style "border-radius" "4px"
                            , style "padding" "16px"
                            , style "margin-bottom" "12px"
                            , style "opacity" "0.7"
                            ]
                            [ h4
                                [ class "mdc-typography--subtitle1"
                                , style "margin-top" "0"
                                , style "text-decoration" "line-through"
                                , style "color" "#9e9e9e"
                                ]
                                [ span
                                    [ style "background-color" "#ef9a9a"
                                    , style "color" "white"
                                    , style "font-size" "11px"
                                    , style "font-weight" "bold"
                                    , style "padding" "1px 6px"
                                    , style "border-radius" "10px"
                                    , style "margin-right" "6px"
                                    ]
                                    [ text "-" ]
                                , text step.name
                                ]
                            , p
                                [ class "mdc-typography--body2"
                                , style "text-decoration" "line-through"
                                , style "color" "#9e9e9e"
                                ]
                                [ text step.instructions ]
                            ]
                        )

            ( _, rendered ) =
                List.foldl
                    (\change ( num, acc ) ->
                        let
                            ( nextNum, el ) =
                                renderChange num change
                        in
                        ( nextNum, acc ++ [ el ] )
                    )
                    ( 1, [] )
                    changes
        in
        div [] rendered


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
