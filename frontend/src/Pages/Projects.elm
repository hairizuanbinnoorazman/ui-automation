module Pages.Projects exposing (Model, Msg, init, update, view)

import API
import Html exposing (Html)
import Html.Attributes
import Html.Events
import Http
import Time
import Types exposing (PaginatedResponse, Project, ProjectInput)



-- MODEL


type alias Model =
    { projects : List Project
    , total : Int
    , limit : Int
    , offset : Int
    , loading : Bool
    , error : Maybe String
    , createDialog : Maybe DialogState
    , editDialog : Maybe EditDialogState
    , deleteDialog : Maybe Project
    , navigationTarget : Maybe String
    }


type alias DialogState =
    { name : String
    , description : String
    }


type alias EditDialogState =
    { project : Project
    , name : String
    , description : String
    }


init : ( Model, Cmd Msg )
init =
    ( { projects = []
      , total = 0
      , limit = 10
      , offset = 0
      , loading = False
      , error = Nothing
      , createDialog = Nothing
      , editDialog = Nothing
      , deleteDialog = Nothing
      , navigationTarget = Nothing
      }
    , API.getProjects 10 0 ProjectsResponse
    )



-- UPDATE


type Msg
    = ProjectsResponse (Result Http.Error (PaginatedResponse Project))
    | LoadPage Int
    | OpenCreateDialog
    | CloseCreateDialog
    | SetCreateName String
    | SetCreateDescription String
    | SubmitCreate
    | CreateResponse (Result Http.Error Project)
    | OpenEditDialog Project
    | CloseEditDialog
    | SetEditName String
    | SetEditDescription String
    | SubmitEdit
    | EditResponse (Result Http.Error Project)
    | OpenDeleteDialog Project
    | CloseDeleteDialog
    | ConfirmDelete String
    | DeleteResponse (Result Http.Error ())
    | NavigateToProcedures String


update : Msg -> Model -> ( Model, Cmd Msg )
update msg model =
    case msg of
        ProjectsResponse (Ok response) ->
            ( { model
                | projects = response.items
                , total = response.total
                , loading = False
                , error = Nothing
              }
            , Cmd.none
            )

        ProjectsResponse (Err error) ->
            ( { model
                | loading = False
                , error = Just (httpErrorToString error)
              }
            , Cmd.none
            )

        LoadPage offset ->
            ( { model | loading = True, offset = offset }
            , API.getProjects model.limit offset ProjectsResponse
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
                    ( { model | loading = True }
                    , API.createProject
                        { name = dialog.name
                        , description = dialog.description
                        }
                        CreateResponse
                    )

                Nothing ->
                    ( model, Cmd.none )

        CreateResponse (Ok project) ->
            ( { model
                | loading = False
                , createDialog = Nothing
                , navigationTarget = Just project.id
              }
            , API.getProjects model.limit model.offset ProjectsResponse
            )

        CreateResponse (Err error) ->
            ( { model
                | loading = False
                , error = Just (httpErrorToString error)
              }
            , Cmd.none
            )

        OpenEditDialog project ->
            ( { model
                | editDialog =
                    Just
                        { project = project
                        , name = project.name
                        , description = project.description
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

        SubmitEdit ->
            case model.editDialog of
                Just dialog ->
                    ( { model | loading = True }
                    , API.updateProject
                        dialog.project.id
                        { name = dialog.name
                        , description = dialog.description
                        }
                        EditResponse
                    )

                Nothing ->
                    ( model, Cmd.none )

        EditResponse (Ok project) ->
            ( { model
                | loading = False
                , editDialog = Nothing
              }
            , API.getProjects model.limit model.offset ProjectsResponse
            )

        EditResponse (Err error) ->
            ( { model
                | loading = False
                , error = Just (httpErrorToString error)
              }
            , Cmd.none
            )

        OpenDeleteDialog project ->
            ( { model | deleteDialog = Just project }
            , Cmd.none
            )

        CloseDeleteDialog ->
            ( { model | deleteDialog = Nothing }
            , Cmd.none
            )

        ConfirmDelete id ->
            ( { model | loading = True }
            , API.deleteProject id DeleteResponse
            )

        DeleteResponse (Ok ()) ->
            ( { model
                | loading = False
                , deleteDialog = Nothing
              }
            , API.getProjects model.limit model.offset ProjectsResponse
            )

        DeleteResponse (Err error) ->
            ( { model
                | loading = False
                , error = Just (httpErrorToString error)
              }
            , Cmd.none
            )

        NavigateToProcedures projectId ->
            ( { model | navigationTarget = Just projectId }
            , Cmd.none
            )



-- VIEW


view : Model -> Html Msg
view model =
    Html.div []
        [ Html.div
            [ Html.Attributes.style "display" "flex"
            , Html.Attributes.style "justify-content" "space-between"
            , Html.Attributes.style "align-items" "center"
            , Html.Attributes.style "margin-bottom" "20px"
            ]
            [ Html.h1 [ Html.Attributes.class "mdc-typography--headline3" ] [ Html.text "Projects" ]
            , Html.button
                [ Html.Events.onClick OpenCreateDialog
                , Html.Attributes.class "mdc-button mdc-button--raised"
                ]
                [ Html.text "Create Project" ]
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
        , if model.loading && List.isEmpty model.projects then
            Html.div [] [ Html.text "Loading..." ]

          else
            viewProjectsTable model.projects
        , viewPagination model
        , viewCreateDialog model.createDialog
        , case model.editDialog of
            Just dialog ->
                viewEditDialog dialog

            Nothing ->
                Html.text ""
        , case model.deleteDialog of
            Just project ->
                viewDeleteDialog project

            Nothing ->
                Html.text ""
        ]


viewProjectsTable : List Project -> Html Msg
viewProjectsTable projects =
    Html.table
        [ Html.Attributes.class "mdc-data-table__table"
        , Html.Attributes.style "width" "100%"
        , Html.Attributes.style "border-collapse" "collapse"
        ]
        [ Html.thead []
            [ Html.tr []
                [ Html.th [ Html.Attributes.style "text-align" "left", Html.Attributes.style "padding" "12px" ] [ Html.text "Name" ]
                , Html.th [ Html.Attributes.style "text-align" "left", Html.Attributes.style "padding" "12px" ] [ Html.text "Description" ]
                , Html.th [ Html.Attributes.style "text-align" "left", Html.Attributes.style "padding" "12px" ] [ Html.text "Created" ]
                , Html.th [ Html.Attributes.style "text-align" "left", Html.Attributes.style "padding" "12px" ] [ Html.text "Actions" ]
                ]
            ]
        , Html.tbody []
            (List.map viewProjectRow projects)
        ]


viewProjectRow : Project -> Html Msg
viewProjectRow project =
    Html.tr [ Html.Attributes.style "border-bottom" "1px solid #ddd" ]
        [ Html.td [ Html.Attributes.style "padding" "12px" ] [ Html.text project.name ]
        , Html.td [ Html.Attributes.style "padding" "12px" ] [ Html.text project.description ]
        , Html.td [ Html.Attributes.style "padding" "12px" ] [ Html.text (formatTime project.createdAt) ]
        , Html.td [ Html.Attributes.style "padding" "12px" ]
            [ Html.button
                [ Html.Events.onClick (NavigateToProcedures project.id)
                , Html.Attributes.class "mdc-button mdc-button--raised"
                , Html.Attributes.style "margin-right" "8px"
                ]
                [ Html.text "View Procedures" ]
            , Html.button
                [ Html.Events.onClick (OpenEditDialog project)
                , Html.Attributes.class "mdc-button"
                , Html.Attributes.style "margin-right" "8px"
                ]
                [ Html.text "Edit" ]
            , Html.button
                [ Html.Events.onClick (OpenDeleteDialog project)
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


viewCreateDialog : Maybe DialogState -> Html Msg
viewCreateDialog maybeDialog =
    case maybeDialog of
        Just dialog ->
            viewDialogOverlay "Create Project"
                [ Html.div [ Html.Attributes.style "margin-bottom" "16px" ]
                    [ Html.label [] [ Html.text "Name" ]
                    , Html.input
                        [ Html.Attributes.type_ "text"
                        , Html.Attributes.value dialog.name
                        , Html.Events.onInput SetCreateName
                        , Html.Attributes.required True
                        , Html.Attributes.style "width" "100%"
                        , Html.Attributes.style "padding" "8px"
                        ]
                        []
                    ]
                , Html.div [ Html.Attributes.style "margin-bottom" "16px" ]
                    [ Html.label [] [ Html.text "Description" ]
                    , Html.input
                        [ Html.Attributes.type_ "text"
                        , Html.Attributes.value dialog.description
                        , Html.Events.onInput SetCreateDescription
                        , Html.Attributes.required True
                        , Html.Attributes.style "width" "100%"
                        , Html.Attributes.style "padding" "8px"
                        ]
                        []
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

        Nothing ->
            Html.text ""


viewEditDialog : EditDialogState -> Html Msg
viewEditDialog dialog =
    viewDialogOverlay "Edit Project"
        [ Html.div [ Html.Attributes.style "margin-bottom" "16px" ]
            [ Html.label [] [ Html.text "Name" ]
            , Html.input
                [ Html.Attributes.type_ "text"
                , Html.Attributes.value dialog.name
                , Html.Events.onInput SetEditName
                , Html.Attributes.required True
                , Html.Attributes.style "width" "100%"
                , Html.Attributes.style "padding" "8px"
                ]
                []
            ]
        , Html.div [ Html.Attributes.style "margin-bottom" "16px" ]
            [ Html.label [] [ Html.text "Description" ]
            , Html.input
                [ Html.Attributes.type_ "text"
                , Html.Attributes.value dialog.description
                , Html.Events.onInput SetEditDescription
                , Html.Attributes.required True
                , Html.Attributes.style "width" "100%"
                , Html.Attributes.style "padding" "8px"
                ]
                []
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


viewDeleteDialog : Project -> Html Msg
viewDeleteDialog project =
    viewDialogOverlay "Delete Project"
        [ Html.div []
            [ Html.text ("Are you sure you want to delete \"" ++ project.name ++ "\"?")
            ]
        ]
        [ Html.button
            [ Html.Events.onClick CloseDeleteDialog
            , Html.Attributes.class "mdc-button"
            ]
            [ Html.text "Cancel" ]
        , Html.button
            [ Html.Events.onClick (ConfirmDelete project.id)
            , Html.Attributes.class "mdc-button mdc-button--raised"
            ]
            [ Html.text "Delete" ]
        ]


viewDialogOverlay : String -> List (Html Msg) -> List (Html Msg) -> Html Msg
viewDialogOverlay title content actions =
    Html.div
        [ Html.Attributes.style "position" "fixed"
        , Html.Attributes.style "top" "0"
        , Html.Attributes.style "left" "0"
        , Html.Attributes.style "width" "100%"
        , Html.Attributes.style "height" "100%"
        , Html.Attributes.style "background-color" "rgba(0,0,0,0.5)"
        , Html.Attributes.style "display" "flex"
        , Html.Attributes.style "justify-content" "center"
        , Html.Attributes.style "align-items" "center"
        , Html.Attributes.style "z-index" "1000"
        ]
        [ Html.div
            [ Html.Attributes.class "mdc-dialog__surface"
            , Html.Attributes.style "background" "white"
            , Html.Attributes.style "padding" "24px"
            , Html.Attributes.style "border-radius" "4px"
            , Html.Attributes.style "min-width" "400px"
            ]
            [ Html.h2 [ Html.Attributes.class "mdc-typography--headline6" ] [ Html.text title ]
            , Html.div [ Html.Attributes.style "margin" "20px 0" ] content
            , Html.div
                [ Html.Attributes.style "display" "flex"
                , Html.Attributes.style "justify-content" "flex-end"
                , Html.Attributes.style "gap" "8px"
                ]
                actions
            ]
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
