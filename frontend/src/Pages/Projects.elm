module Pages.Projects exposing (Model, Msg, init, update, view)

import API
import Html exposing (Html)
import Html.Attributes
import Http
import Material.Button as Button
import Material.Card as Card
import Material.DataTable as DataTable
import Material.Dialog as Dialog
import Material.IconButton as IconButton
import Material.LayoutGrid as LayoutGrid
import Material.TextField as TextField
import Material.Typography as Typography
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
    , createDialog : DialogState
    , editDialog : Maybe EditDialogState
    , deleteDialog : Maybe Project
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
      , createDialog = { name = "", description = "" }
      , editDialog = Nothing
      , deleteDialog = Nothing
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
            ( { model | createDialog = { name = "", description = "" } }
            , Cmd.none
            )

        CloseCreateDialog ->
            ( { model | createDialog = { name = "", description = "" } }
            , Cmd.none
            )

        SetCreateName name ->
            let
                dialog =
                    model.createDialog
            in
            ( { model | createDialog = { dialog | name = name } }
            , Cmd.none
            )

        SetCreateDescription description ->
            let
                dialog =
                    model.createDialog
            in
            ( { model | createDialog = { dialog | description = description } }
            , Cmd.none
            )

        SubmitCreate ->
            ( { model | loading = True }
            , API.createProject
                { name = model.createDialog.name
                , description = model.createDialog.description
                }
                CreateResponse
            )

        CreateResponse (Ok project) ->
            ( { model
                | loading = False
                , createDialog = { name = "", description = "" }
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
                    [ Html.h1 [ Typography.headline3 ] [ Html.text "Projects" ]
                    , Button.raised
                        (Button.config |> Button.setOnClick (Just OpenCreateDialog))
                        "Create Project"
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
                [ if model.loading && List.isEmpty model.projects then
                    Html.div [] [ Html.text "Loading..." ]

                  else
                    viewProjectsTable model.projects
                ]
            , LayoutGrid.cell
                [ LayoutGrid.span12 ]
                [ viewPagination model ]
            ]
        , viewCreateDialog model
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
    DataTable.dataTable
        (DataTable.config |> DataTable.setAttributes [ Typography.typography ])
        { thead =
            [ DataTable.row []
                [ DataTable.cell [] [ Html.text "Name" ]
                , DataTable.cell [] [ Html.text "Description" ]
                , DataTable.cell [] [ Html.text "Created" ]
                , DataTable.cell [] [ Html.text "Actions" ]
                ]
            ]
        , tbody =
            List.map viewProjectRow projects
        }


viewProjectRow : Project -> DataTable.Row Msg
viewProjectRow project =
    DataTable.row []
        [ DataTable.cell [] [ Html.text project.name ]
        , DataTable.cell [] [ Html.text project.description ]
        , DataTable.cell [] [ Html.text (formatTime project.createdAt) ]
        , DataTable.cell []
            [ Button.text
                (Button.config |> Button.setOnClick (Just (OpenEditDialog project)))
                "Edit"
            , Button.text
                (Button.config |> Button.setOnClick (Just (OpenDeleteDialog project)))
                "Delete"
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


viewCreateDialog : Model -> Html Msg
viewCreateDialog model =
    Dialog.dialog
        (Dialog.config
            |> Dialog.setOpen (not (String.isEmpty model.createDialog.name && String.isEmpty model.createDialog.description))
            |> Dialog.setOnClose CloseCreateDialog
        )
        { title = Just "Create Project"
        , content =
            [ Html.div []
                [ TextField.filled
                    (TextField.config
                        |> TextField.setLabel (Just "Name")
                        |> TextField.setValue (Just model.createDialog.name)
                        |> TextField.setOnInput (Just SetCreateName)
                        |> TextField.setRequired True
                    )
                , TextField.filled
                    (TextField.config
                        |> TextField.setLabel (Just "Description")
                        |> TextField.setValue (Just model.createDialog.description)
                        |> TextField.setOnInput (Just SetCreateDescription)
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
        { title = Just "Edit Project"
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


viewDeleteDialog : Project -> Html Msg
viewDeleteDialog project =
    Dialog.dialog
        (Dialog.config
            |> Dialog.setOpen True
            |> Dialog.setOnClose CloseDeleteDialog
        )
        { title = Just "Delete Project"
        , content =
            [ Html.div []
                [ Html.text ("Are you sure you want to delete \"" ++ project.name ++ "\"?")
                ]
            ]
        , actions =
            [ Button.text
                (Button.config |> Button.setOnClick (Just CloseDeleteDialog))
                "Cancel"
            , Button.raised
                (Button.config |> Button.setOnClick (Just (ConfirmDelete project.id)))
                "Delete"
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
