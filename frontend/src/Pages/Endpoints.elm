module Pages.Endpoints exposing (Model, Msg, init, update, view)

import API
import Components
import Html exposing (Html)
import Html.Attributes
import Html.Events
import Http
import Time
import Types exposing (Credential, Endpoint, EndpointInput, PaginatedResponse)



-- MODEL


type alias Model =
    { endpoints : List Endpoint
    , total : Int
    , limit : Int
    , offset : Int
    , loading : Bool
    , error : Maybe String
    , createDialog : Maybe CreateDialogState
    , editDialog : Maybe EditDialogState
    , deleteDialog : Maybe Endpoint
    }


type alias CreateDialogState =
    { name : String
    , url : String
    , credentials : List Credential
    }


type alias EditDialogState =
    { endpoint : Endpoint
    , name : String
    , url : String
    , credentials : List Credential
    }


init : ( Model, Cmd Msg )
init =
    ( { endpoints = []
      , total = 0
      , limit = 10
      , offset = 0
      , loading = True
      , error = Nothing
      , createDialog = Nothing
      , editDialog = Nothing
      , deleteDialog = Nothing
      }
    , API.getEndpoints 10 0 EndpointsResponse
    )



-- UPDATE


type Msg
    = EndpointsResponse (Result Http.Error (PaginatedResponse Endpoint))
    | LoadPage Int
    | OpenCreateDialog
    | CloseCreateDialog
    | SetCreateName String
    | SetCreateUrl String
    | AddCreateCredential
    | RemoveCreateCredential Int
    | SetCreateCredentialKey Int String
    | SetCreateCredentialValue Int String
    | SubmitCreate
    | CreateResponse (Result Http.Error Endpoint)
    | OpenEditDialog Endpoint
    | CloseEditDialog
    | SetEditName String
    | SetEditUrl String
    | AddEditCredential
    | RemoveEditCredential Int
    | SetEditCredentialKey Int String
    | SetEditCredentialValue Int String
    | SubmitEdit
    | EditResponse (Result Http.Error Endpoint)
    | OpenDeleteDialog Endpoint
    | CloseDeleteDialog
    | ConfirmDelete String
    | DeleteResponse (Result Http.Error ())


update : Msg -> Model -> ( Model, Cmd Msg )
update msg model =
    case msg of
        EndpointsResponse (Ok response) ->
            ( { model
                | endpoints = response.items
                , total = response.total
                , loading = False
                , error = Nothing
              }
            , Cmd.none
            )

        EndpointsResponse (Err error) ->
            ( { model
                | loading = False
                , error = Just (httpErrorToString error)
              }
            , Cmd.none
            )

        LoadPage offset ->
            ( { model | loading = True, offset = offset }
            , API.getEndpoints model.limit offset EndpointsResponse
            )

        OpenCreateDialog ->
            ( { model
                | createDialog =
                    Just
                        { name = ""
                        , url = ""
                        , credentials =
                            [ { key = "username", value = "" }
                            , { key = "email", value = "" }
                            ]
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

        SetCreateUrl url ->
            case model.createDialog of
                Just dialog ->
                    ( { model | createDialog = Just { dialog | url = url } }
                    , Cmd.none
                    )

                Nothing ->
                    ( model, Cmd.none )

        AddCreateCredential ->
            case model.createDialog of
                Just dialog ->
                    ( { model | createDialog = Just { dialog | credentials = dialog.credentials ++ [ { key = "", value = "" } ] } }
                    , Cmd.none
                    )

                Nothing ->
                    ( model, Cmd.none )

        RemoveCreateCredential index ->
            case model.createDialog of
                Just dialog ->
                    ( { model | createDialog = Just { dialog | credentials = removeAt index dialog.credentials } }
                    , Cmd.none
                    )

                Nothing ->
                    ( model, Cmd.none )

        SetCreateCredentialKey index key ->
            case model.createDialog of
                Just dialog ->
                    ( { model | createDialog = Just { dialog | credentials = updateCredentialKey index key dialog.credentials } }
                    , Cmd.none
                    )

                Nothing ->
                    ( model, Cmd.none )

        SetCreateCredentialValue index value ->
            case model.createDialog of
                Just dialog ->
                    ( { model | createDialog = Just { dialog | credentials = updateCredentialValue index value dialog.credentials } }
                    , Cmd.none
                    )

                Nothing ->
                    ( model, Cmd.none )

        SubmitCreate ->
            case model.createDialog of
                Just dialog ->
                    ( { model | loading = True }
                    , API.createEndpoint
                        { name = dialog.name
                        , url = dialog.url
                        , credentials = dialog.credentials
                        }
                        CreateResponse
                    )

                Nothing ->
                    ( model, Cmd.none )

        CreateResponse (Ok _) ->
            ( { model
                | loading = False
                , createDialog = Nothing
              }
            , API.getEndpoints model.limit model.offset EndpointsResponse
            )

        CreateResponse (Err error) ->
            ( { model
                | loading = False
                , error = Just (httpErrorToString error)
              }
            , Cmd.none
            )

        OpenEditDialog endpoint ->
            ( { model
                | editDialog =
                    Just
                        { endpoint = endpoint
                        , name = endpoint.name
                        , url = endpoint.url
                        , credentials = endpoint.credentials
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

        SetEditUrl url ->
            case model.editDialog of
                Just dialog ->
                    ( { model | editDialog = Just { dialog | url = url } }
                    , Cmd.none
                    )

                Nothing ->
                    ( model, Cmd.none )

        AddEditCredential ->
            case model.editDialog of
                Just dialog ->
                    ( { model | editDialog = Just { dialog | credentials = dialog.credentials ++ [ { key = "", value = "" } ] } }
                    , Cmd.none
                    )

                Nothing ->
                    ( model, Cmd.none )

        RemoveEditCredential index ->
            case model.editDialog of
                Just dialog ->
                    ( { model | editDialog = Just { dialog | credentials = removeAt index dialog.credentials } }
                    , Cmd.none
                    )

                Nothing ->
                    ( model, Cmd.none )

        SetEditCredentialKey index key ->
            case model.editDialog of
                Just dialog ->
                    ( { model | editDialog = Just { dialog | credentials = updateCredentialKey index key dialog.credentials } }
                    , Cmd.none
                    )

                Nothing ->
                    ( model, Cmd.none )

        SetEditCredentialValue index value ->
            case model.editDialog of
                Just dialog ->
                    ( { model | editDialog = Just { dialog | credentials = updateCredentialValue index value dialog.credentials } }
                    , Cmd.none
                    )

                Nothing ->
                    ( model, Cmd.none )

        SubmitEdit ->
            case model.editDialog of
                Just dialog ->
                    ( { model | loading = True }
                    , API.updateEndpoint
                        dialog.endpoint.id
                        { name = dialog.name
                        , url = dialog.url
                        , credentials = dialog.credentials
                        }
                        EditResponse
                    )

                Nothing ->
                    ( model, Cmd.none )

        EditResponse (Ok _) ->
            ( { model
                | loading = False
                , editDialog = Nothing
              }
            , API.getEndpoints model.limit model.offset EndpointsResponse
            )

        EditResponse (Err error) ->
            ( { model
                | loading = False
                , error = Just (httpErrorToString error)
              }
            , Cmd.none
            )

        OpenDeleteDialog endpoint ->
            ( { model | deleteDialog = Just endpoint }
            , Cmd.none
            )

        CloseDeleteDialog ->
            ( { model | deleteDialog = Nothing }
            , Cmd.none
            )

        ConfirmDelete id ->
            ( { model | loading = True }
            , API.deleteEndpoint id DeleteResponse
            )

        DeleteResponse (Ok ()) ->
            ( { model
                | loading = False
                , deleteDialog = Nothing
              }
            , API.getEndpoints model.limit model.offset EndpointsResponse
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
        [ Html.div
            [ Html.Attributes.class "page-header"
            ]
            [ Html.h1 [ Html.Attributes.class "mdc-typography--headline3" ] [ Html.text "Endpoints" ]
            , Html.button
                [ Html.Events.onClick OpenCreateDialog
                , Html.Attributes.class "mdc-button mdc-button--raised"
                ]
                [ Html.text "Create Endpoint" ]
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
        , if model.loading && List.isEmpty model.endpoints then
            Html.div [] [ Html.text "Loading..." ]

          else
            viewEndpointsTable model.endpoints
        , viewPagination model
        , viewCreateDialog model.createDialog
        , case model.editDialog of
            Just dialog ->
                viewEditDialog dialog

            Nothing ->
                Html.text ""
        , case model.deleteDialog of
            Just endpoint ->
                viewDeleteDialog endpoint

            Nothing ->
                Html.text ""
        ]


viewEndpointsTable : List Endpoint -> Html Msg
viewEndpointsTable endpoints =
    Html.table
        [ Html.Attributes.class "mdc-data-table__table"
        , Html.Attributes.style "width" "100%"
        , Html.Attributes.style "border-collapse" "collapse"
        ]
        [ Html.thead []
            [ Html.tr []
                [ Html.th [ Html.Attributes.style "text-align" "left", Html.Attributes.style "padding" "12px" ] [ Html.text "Name" ]
                , Html.th [ Html.Attributes.style "text-align" "left", Html.Attributes.style "padding" "12px" ] [ Html.text "URL" ]
                , Html.th [ Html.Attributes.style "text-align" "left", Html.Attributes.style "padding" "12px" ] [ Html.text "Credentials Count" ]
                , Html.th [ Html.Attributes.style "text-align" "left", Html.Attributes.style "padding" "12px" ] [ Html.text "Created" ]
                , Html.th [ Html.Attributes.style "text-align" "left", Html.Attributes.style "padding" "12px" ] [ Html.text "Actions" ]
                ]
            ]
        , Html.tbody []
            (List.map viewEndpointRow endpoints)
        ]


viewEndpointRow : Endpoint -> Html Msg
viewEndpointRow endpoint =
    Html.tr [ Html.Attributes.style "border-bottom" "1px solid #ddd" ]
        [ Html.td [ Html.Attributes.style "padding" "12px" ] [ Html.text endpoint.name ]
        , Html.td [ Html.Attributes.style "padding" "12px" ] [ Html.text endpoint.url ]
        , Html.td [ Html.Attributes.style "padding" "12px" ] [ Html.text (String.fromInt (List.length endpoint.credentials)) ]
        , Html.td [ Html.Attributes.style "padding" "12px" ] [ Html.text (formatTime endpoint.createdAt) ]
        , Html.td [ Html.Attributes.style "padding" "12px" ]
            [ Html.button
                [ Html.Events.onClick (OpenEditDialog endpoint)
                , Html.Attributes.class "mdc-button"
                , Html.Attributes.style "margin-right" "8px"
                ]
                [ Html.text "Edit" ]
            , Html.button
                [ Html.Events.onClick (OpenDeleteDialog endpoint)
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


viewCredentialWarning : Html msg
viewCredentialWarning =
    Html.div
        [ Html.Attributes.style "background-color" "#fff3cd"
        , Html.Attributes.style "border" "1px solid #ffc107"
        , Html.Attributes.style "border-radius" "4px"
        , Html.Attributes.style "padding" "12px"
        , Html.Attributes.style "margin-bottom" "16px"
        , Html.Attributes.style "color" "#856404"
        , Html.Attributes.style "font-size" "13px"
        ]
        [ Html.text "Credentials are stored as plain text. Only use for testing purposes." ]


viewCredentialRows : (Int -> String -> msg) -> (Int -> String -> msg) -> (Int -> msg) -> msg -> List Credential -> List (Html msg)
viewCredentialRows onKeyChange onValueChange onRemove onAdd credentials =
    [ Html.div
        [ Html.Attributes.style "margin-bottom" "12px" ]
        [ Html.label
            [ Html.Attributes.style "display" "block"
            , Html.Attributes.style "margin-bottom" "8px"
            , Html.Attributes.style "font-weight" "500"
            , Html.Attributes.style "color" "#333"
            ]
            [ Html.text "Credentials" ]
        , Html.div []
            (List.indexedMap
                (\index cred ->
                    Html.div
                        [ Html.Attributes.style "display" "flex"
                        , Html.Attributes.style "gap" "8px"
                        , Html.Attributes.style "margin-bottom" "8px"
                        , Html.Attributes.style "align-items" "center"
                        ]
                        [ Html.input
                            [ Html.Attributes.type_ "text"
                            , Html.Attributes.placeholder "Key"
                            , Html.Attributes.value cred.key
                            , Html.Events.onInput (onKeyChange index)
                            , Html.Attributes.style "flex" "1"
                            , Html.Attributes.style "padding" "8px"
                            , Html.Attributes.style "border" "1px solid #ddd"
                            , Html.Attributes.style "border-radius" "4px"
                            , Html.Attributes.style "font-size" "14px"
                            ]
                            []
                        , Html.input
                            [ Html.Attributes.type_ "text"
                            , Html.Attributes.placeholder "Value"
                            , Html.Attributes.value cred.value
                            , Html.Events.onInput (onValueChange index)
                            , Html.Attributes.style "flex" "1"
                            , Html.Attributes.style "padding" "8px"
                            , Html.Attributes.style "border" "1px solid #ddd"
                            , Html.Attributes.style "border-radius" "4px"
                            , Html.Attributes.style "font-size" "14px"
                            ]
                            []
                        , Html.button
                            [ Html.Events.onClick (onRemove index)
                            , Html.Attributes.style "padding" "6px 12px"
                            , Html.Attributes.style "background" "#f44336"
                            , Html.Attributes.style "color" "white"
                            , Html.Attributes.style "border" "none"
                            , Html.Attributes.style "border-radius" "4px"
                            , Html.Attributes.style "cursor" "pointer"
                            , Html.Attributes.style "font-size" "14px"
                            ]
                            [ Html.text "Remove" ]
                        ]
                )
                credentials
            )
        , Html.button
            [ Html.Events.onClick onAdd
            , Html.Attributes.class "mdc-button"
            , Html.Attributes.style "margin-top" "4px"
            ]
            [ Html.text "Add Credential" ]
        ]
    ]


viewCreateDialog : Maybe CreateDialogState -> Html Msg
viewCreateDialog maybeDialog =
    case maybeDialog of
        Just dialog ->
            Components.viewDialogOverlay "Create Endpoint"
                ([ viewCredentialWarning
                 , Components.viewFormField "Name"
                    [ Html.Attributes.type_ "text"
                    , Html.Attributes.value dialog.name
                    , Html.Events.onInput SetCreateName
                    , Html.Attributes.required True
                    ]
                 , Components.viewFormField "URL"
                    [ Html.Attributes.type_ "text"
                    , Html.Attributes.value dialog.url
                    , Html.Events.onInput SetCreateUrl
                    , Html.Attributes.required True
                    ]
                 ]
                    ++ viewCredentialRows SetCreateCredentialKey SetCreateCredentialValue RemoveCreateCredential AddCreateCredential dialog.credentials
                )
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
    Components.viewDialogOverlay "Edit Endpoint"
        ([ viewCredentialWarning
         , Components.viewFormField "Name"
            [ Html.Attributes.type_ "text"
            , Html.Attributes.value dialog.name
            , Html.Events.onInput SetEditName
            , Html.Attributes.required True
            ]
         , Components.viewFormField "URL"
            [ Html.Attributes.type_ "text"
            , Html.Attributes.value dialog.url
            , Html.Events.onInput SetEditUrl
            , Html.Attributes.required True
            ]
         ]
            ++ viewCredentialRows SetEditCredentialKey SetEditCredentialValue RemoveEditCredential AddEditCredential dialog.credentials
        )
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


viewDeleteDialog : Endpoint -> Html Msg
viewDeleteDialog endpoint =
    Components.viewDialogOverlay "Delete Endpoint"
        [ Html.div []
            [ Html.text ("Are you sure you want to delete \"" ++ endpoint.name ++ "\"?")
            ]
        ]
        [ Html.button
            [ Html.Events.onClick CloseDeleteDialog
            , Html.Attributes.class "mdc-button"
            ]
            [ Html.text "Cancel" ]
        , Html.button
            [ Html.Events.onClick (ConfirmDelete endpoint.id)
            , Html.Attributes.class "mdc-button mdc-button--raised"
            ]
            [ Html.text "Delete" ]
        ]



-- HELPERS


removeAt : Int -> List a -> List a
removeAt index list =
    List.take index list ++ List.drop (index + 1) list


updateCredentialKey : Int -> String -> List Credential -> List Credential
updateCredentialKey index newKey credentials =
    List.indexedMap
        (\i cred ->
            if i == index then
                { cred | key = newKey }

            else
                cred
        )
        credentials


updateCredentialValue : Int -> String -> List Credential -> List Credential
updateCredentialValue index newValue credentials =
    List.indexedMap
        (\i cred ->
            if i == index then
                { cred | value = newValue }

            else
                cred
        )
        credentials


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
