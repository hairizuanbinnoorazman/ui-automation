module Pages.TestProcedures exposing (Model, Msg, init, update, view)

import API
import Components
import Html exposing (Html, button, div, h1, p, span, text)
import Html.Attributes exposing (class, disabled, placeholder, style, type_, value)
import Html.Events exposing (onClick, onInput)
import Http
import Types exposing (PaginatedResponse, TestProcedure)


-- MODEL


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
    , navigationTarget : Maybe String
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
      , navigationTarget = Nothing
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
            ( { model | navigationTarget = Just procedure.id }, Cmd.none )



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
        , viewProcedureList model
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
                        , Html.th [ style "text-align" "left", style "padding" "12px" ] [ text "Description" ]
                        , Html.th [ style "text-align" "left", style "padding" "12px" ] [ text "Version" ]
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
                                , Html.td [ style "padding" "12px" ] [ text proc.description ]
                                , Html.td [ style "padding" "12px" ] [ text (String.fromInt proc.version) ]
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
