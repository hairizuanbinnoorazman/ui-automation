module Pages.Manage exposing (Model, Msg, init, update, view)

import API
import Components
import Html exposing (Html)
import Html.Attributes
import Html.Events
import Http
import Types exposing (APIToken, CreateTokenInput, CreateTokenResponse, Credential, Integration, IntegrationListResponse, TokenListResponse)



-- MODEL


type alias CreateIntegrationDialogState =
    { name : String
    , provider : String
    , jiraUrl : String
    , jiraEmail : String
    , jiraApiToken : String
    , jiraDefaultProject : String
    , githubToken : String
    , githubDefaultOwner : String
    , githubDefaultRepo : String
    }


type alias Model =
    { tokens : List APIToken
    , total : Int
    , loading : Bool
    , error : Maybe String
    , createDialog : Maybe CreateDialogState
    , createdTokenSecret : Maybe String
    , deleteDialog : Maybe APIToken
    , integrations : List Integration
    , integrationsLoading : Bool
    , createIntegrationDialog : Maybe CreateIntegrationDialogState
    , deleteIntegrationDialog : Maybe Integration
    , testConnectionResult : Maybe (Result String String)
    }


type alias CreateDialogState =
    { name : String
    , scope : String
    , expiresInHours : Int
    }


init : ( Model, Cmd Msg )
init =
    ( { tokens = []
      , total = 0
      , loading = True
      , error = Nothing
      , createDialog = Nothing
      , createdTokenSecret = Nothing
      , deleteDialog = Nothing
      , integrations = []
      , integrationsLoading = True
      , createIntegrationDialog = Nothing
      , deleteIntegrationDialog = Nothing
      , testConnectionResult = Nothing
      }
    , Cmd.batch
        [ API.getAPITokens TokensResponse
        , API.getIntegrations IntegrationsResponse
        ]
    )



-- UPDATE


type Msg
    = TokensResponse (Result Http.Error TokenListResponse)
    | OpenCreateDialog
    | CloseCreateDialog
    | SetCreateName String
    | SetCreateScope String
    | SetCreateExpiry String
    | SubmitCreate
    | CreateResponse (Result Http.Error CreateTokenResponse)
    | DismissTokenSecret
    | OpenDeleteDialog APIToken
    | CloseDeleteDialog
    | ConfirmDelete String
    | DeleteResponse (Result Http.Error ())
    | IntegrationsResponse (Result Http.Error IntegrationListResponse)
    | OpenCreateIntegrationDialog
    | CloseCreateIntegrationDialog
    | SetIntegrationName String
    | SetIntegrationProvider String
    | SetJiraUrl String
    | SetJiraEmail String
    | SetJiraApiToken String
    | SetJiraDefaultProject String
    | SetGithubToken String
    | SetGithubDefaultOwner String
    | SetGithubDefaultRepo String
    | SubmitCreateIntegration
    | CreateIntegrationResponse (Result Http.Error Integration)
    | OpenDeleteIntegrationDialog Integration
    | CloseDeleteIntegrationDialog
    | ConfirmDeleteIntegration String
    | DeleteIntegrationResponse (Result Http.Error ())
    | TestConnection String
    | TestConnectionResponse (Result Http.Error ())


update : Msg -> Model -> ( Model, Cmd Msg )
update msg model =
    case msg of
        TokensResponse (Ok response) ->
            ( { model
                | tokens = response.tokens
                , total = response.total
                , loading = False
                , error = Nothing
              }
            , Cmd.none
            )

        TokensResponse (Err error) ->
            ( { model
                | loading = False
                , error = Just (httpErrorToString error)
              }
            , Cmd.none
            )

        OpenCreateDialog ->
            ( { model
                | createDialog =
                    Just
                        { name = ""
                        , scope = "read_only"
                        , expiresInHours = 720
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

        SetCreateScope scope ->
            case model.createDialog of
                Just dialog ->
                    ( { model | createDialog = Just { dialog | scope = scope } }
                    , Cmd.none
                    )

                Nothing ->
                    ( model, Cmd.none )

        SetCreateExpiry expiryStr ->
            case model.createDialog of
                Just dialog ->
                    let
                        hours =
                            Maybe.withDefault 720 (String.toInt expiryStr)
                    in
                    ( { model | createDialog = Just { dialog | expiresInHours = hours } }
                    , Cmd.none
                    )

                Nothing ->
                    ( model, Cmd.none )

        SubmitCreate ->
            case model.createDialog of
                Just dialog ->
                    ( { model | loading = True }
                    , API.createAPIToken
                        { name = dialog.name
                        , scope = dialog.scope
                        , expiresInHours = dialog.expiresInHours
                        }
                        CreateResponse
                    )

                Nothing ->
                    ( model, Cmd.none )

        CreateResponse (Ok response) ->
            ( { model
                | loading = False
                , createDialog = Nothing
                , createdTokenSecret = Just response.token
              }
            , API.getAPITokens TokensResponse
            )

        CreateResponse (Err error) ->
            ( { model
                | loading = False
                , error = Just (httpErrorToString error)
              }
            , Cmd.none
            )

        DismissTokenSecret ->
            ( { model | createdTokenSecret = Nothing }
            , Cmd.none
            )

        OpenDeleteDialog token ->
            ( { model | deleteDialog = Just token }
            , Cmd.none
            )

        CloseDeleteDialog ->
            ( { model | deleteDialog = Nothing }
            , Cmd.none
            )

        ConfirmDelete tokenId ->
            ( { model | loading = True }
            , API.revokeAPIToken tokenId DeleteResponse
            )

        DeleteResponse (Ok ()) ->
            ( { model
                | loading = False
                , deleteDialog = Nothing
              }
            , API.getAPITokens TokensResponse
            )

        DeleteResponse (Err error) ->
            ( { model
                | loading = False
                , error = Just (httpErrorToString error)
              }
            , Cmd.none
            )

        IntegrationsResponse (Ok response) ->
            ( { model
                | integrations = response.items
                , integrationsLoading = False
              }
            , Cmd.none
            )

        IntegrationsResponse (Err error) ->
            ( { model
                | integrationsLoading = False
                , error = Just (httpErrorToString error)
              }
            , Cmd.none
            )

        OpenCreateIntegrationDialog ->
            ( { model
                | createIntegrationDialog =
                    Just
                        { name = ""
                        , provider = "jira"
                        , jiraUrl = ""
                        , jiraEmail = ""
                        , jiraApiToken = ""
                        , jiraDefaultProject = ""
                        , githubToken = ""
                        , githubDefaultOwner = ""
                        , githubDefaultRepo = ""
                        }
              }
            , Cmd.none
            )

        CloseCreateIntegrationDialog ->
            ( { model | createIntegrationDialog = Nothing }
            , Cmd.none
            )

        SetIntegrationName name ->
            case model.createIntegrationDialog of
                Just dialog ->
                    ( { model | createIntegrationDialog = Just { dialog | name = name } }
                    , Cmd.none
                    )

                Nothing ->
                    ( model, Cmd.none )

        SetIntegrationProvider provider ->
            case model.createIntegrationDialog of
                Just dialog ->
                    ( { model | createIntegrationDialog = Just { dialog | provider = provider } }
                    , Cmd.none
                    )

                Nothing ->
                    ( model, Cmd.none )

        SetJiraUrl url ->
            case model.createIntegrationDialog of
                Just dialog ->
                    ( { model | createIntegrationDialog = Just { dialog | jiraUrl = url } }
                    , Cmd.none
                    )

                Nothing ->
                    ( model, Cmd.none )

        SetJiraEmail email ->
            case model.createIntegrationDialog of
                Just dialog ->
                    ( { model | createIntegrationDialog = Just { dialog | jiraEmail = email } }
                    , Cmd.none
                    )

                Nothing ->
                    ( model, Cmd.none )

        SetJiraApiToken token ->
            case model.createIntegrationDialog of
                Just dialog ->
                    ( { model | createIntegrationDialog = Just { dialog | jiraApiToken = token } }
                    , Cmd.none
                    )

                Nothing ->
                    ( model, Cmd.none )

        SetJiraDefaultProject project ->
            case model.createIntegrationDialog of
                Just dialog ->
                    ( { model | createIntegrationDialog = Just { dialog | jiraDefaultProject = project } }
                    , Cmd.none
                    )

                Nothing ->
                    ( model, Cmd.none )

        SetGithubToken token ->
            case model.createIntegrationDialog of
                Just dialog ->
                    ( { model | createIntegrationDialog = Just { dialog | githubToken = token } }
                    , Cmd.none
                    )

                Nothing ->
                    ( model, Cmd.none )

        SetGithubDefaultOwner owner ->
            case model.createIntegrationDialog of
                Just dialog ->
                    ( { model | createIntegrationDialog = Just { dialog | githubDefaultOwner = owner } }
                    , Cmd.none
                    )

                Nothing ->
                    ( model, Cmd.none )

        SetGithubDefaultRepo repo ->
            case model.createIntegrationDialog of
                Just dialog ->
                    ( { model | createIntegrationDialog = Just { dialog | githubDefaultRepo = repo } }
                    , Cmd.none
                    )

                Nothing ->
                    ( model, Cmd.none )

        SubmitCreateIntegration ->
            case model.createIntegrationDialog of
                Just dialog ->
                    let
                        credentials =
                            if dialog.provider == "jira" then
                                [ Credential "url" dialog.jiraUrl
                                , Credential "email" dialog.jiraEmail
                                , Credential "api_token" dialog.jiraApiToken
                                , Credential "default_project" dialog.jiraDefaultProject
                                ]

                            else
                                [ Credential "token" dialog.githubToken
                                , Credential "default_owner" dialog.githubDefaultOwner
                                , Credential "default_repo" dialog.githubDefaultRepo
                                ]
                    in
                    ( { model | integrationsLoading = True }
                    , API.createIntegration
                        { name = dialog.name
                        , provider = dialog.provider
                        , credentials = credentials
                        }
                        CreateIntegrationResponse
                    )

                Nothing ->
                    ( model, Cmd.none )

        CreateIntegrationResponse (Ok _) ->
            ( { model
                | integrationsLoading = False
                , createIntegrationDialog = Nothing
              }
            , API.getIntegrations IntegrationsResponse
            )

        CreateIntegrationResponse (Err error) ->
            ( { model
                | integrationsLoading = False
                , error = Just (httpErrorToString error)
              }
            , Cmd.none
            )

        OpenDeleteIntegrationDialog integration ->
            ( { model | deleteIntegrationDialog = Just integration }
            , Cmd.none
            )

        CloseDeleteIntegrationDialog ->
            ( { model | deleteIntegrationDialog = Nothing }
            , Cmd.none
            )

        ConfirmDeleteIntegration integrationId ->
            ( { model | integrationsLoading = True }
            , API.deleteIntegration integrationId DeleteIntegrationResponse
            )

        DeleteIntegrationResponse (Ok ()) ->
            ( { model
                | integrationsLoading = False
                , deleteIntegrationDialog = Nothing
              }
            , API.getIntegrations IntegrationsResponse
            )

        DeleteIntegrationResponse (Err error) ->
            ( { model
                | integrationsLoading = False
                , error = Just (httpErrorToString error)
              }
            , Cmd.none
            )

        TestConnection integrationId ->
            ( { model | testConnectionResult = Nothing }
            , API.testIntegrationConnection integrationId TestConnectionResponse
            )

        TestConnectionResponse (Ok ()) ->
            ( { model | testConnectionResult = Just (Ok "Connection successful") }
            , Cmd.none
            )

        TestConnectionResponse (Err error) ->
            ( { model | testConnectionResult = Just (Err (httpErrorToString error)) }
            , Cmd.none
            )



-- VIEW


view : Model -> Html Msg
view model =
    Html.div []
        [ Html.div
            [ Html.Attributes.class "page-header"
            ]
            [ Html.h1 [ Html.Attributes.class "mdc-typography--headline3" ] [ Html.text "Account Management" ]
            ]
        , Html.h2
            [ Html.Attributes.class "mdc-typography--headline5"
            , Html.Attributes.style "margin-bottom" "16px"
            ]
            [ Html.text "API Tokens" ]
        , Html.div
            [ Html.Attributes.style "margin-bottom" "20px" ]
            [ Html.button
                [ Html.Events.onClick OpenCreateDialog
                , Html.Attributes.class "mdc-button mdc-button--raised"
                ]
                [ Html.text "Create Token" ]
            ]
        , case model.createdTokenSecret of
            Just secret ->
                viewTokenSecretBanner secret

            Nothing ->
                Html.text ""
        , case model.error of
            Just err ->
                Html.div
                    [ Html.Attributes.style "color" "red"
                    , Html.Attributes.style "margin-bottom" "20px"
                    ]
                    [ Html.text err ]

            Nothing ->
                Html.text ""
        , if model.loading && List.isEmpty model.tokens then
            Html.div [] [ Html.text "Loading..." ]

          else
            viewTokensTable model.tokens
        , viewCreateDialog model.createDialog
        , case model.deleteDialog of
            Just token ->
                viewDeleteDialog token

            Nothing ->
                Html.text ""
        , Html.hr [ Html.Attributes.style "margin" "40px 0" ] []
        , viewIntegrationsSection model
        , viewCreateIntegrationDialog model.createIntegrationDialog
        , case model.deleteIntegrationDialog of
            Just integration ->
                viewDeleteIntegrationDialog integration

            Nothing ->
                Html.text ""
        ]


viewTokenSecretBanner : String -> Html Msg
viewTokenSecretBanner secret =
    Html.div
        [ Html.Attributes.style "background-color" "#e8f5e9"
        , Html.Attributes.style "border" "2px solid #4caf50"
        , Html.Attributes.style "border-radius" "4px"
        , Html.Attributes.style "padding" "20px"
        , Html.Attributes.style "margin-bottom" "20px"
        ]
        [ Html.div
            [ Html.Attributes.style "font-weight" "bold"
            , Html.Attributes.style "margin-bottom" "12px"
            , Html.Attributes.style "color" "#2e7d32"
            , Html.Attributes.style "font-size" "16px"
            ]
            [ Html.text "Token created successfully! Copy it now - it won't be shown again." ]
        , Html.code
            [ Html.Attributes.style "display" "block"
            , Html.Attributes.style "background" "#f5f5f5"
            , Html.Attributes.style "padding" "12px"
            , Html.Attributes.style "border-radius" "4px"
            , Html.Attributes.style "font-size" "14px"
            , Html.Attributes.style "word-break" "break-all"
            , Html.Attributes.style "user-select" "all"
            , Html.Attributes.style "margin-bottom" "12px"
            ]
            [ Html.text secret ]
        , Html.button
            [ Html.Events.onClick DismissTokenSecret
            , Html.Attributes.class "mdc-button"
            ]
            [ Html.text "Dismiss" ]
        ]


viewTokensTable : List APIToken -> Html Msg
viewTokensTable tokens =
    if List.isEmpty tokens then
        Html.div
            [ Html.Attributes.style "color" "#666"
            , Html.Attributes.style "padding" "20px"
            ]
            [ Html.text "No API tokens yet. Create one to get started." ]

    else
        Html.table
            [ Html.Attributes.class "mdc-data-table__table"
            , Html.Attributes.style "width" "100%"
            , Html.Attributes.style "border-collapse" "collapse"
            ]
            [ Html.thead []
                [ Html.tr []
                    [ Html.th [ Html.Attributes.style "text-align" "left", Html.Attributes.style "padding" "12px" ] [ Html.text "Name" ]
                    , Html.th [ Html.Attributes.style "text-align" "left", Html.Attributes.style "padding" "12px" ] [ Html.text "Scope" ]
                    , Html.th [ Html.Attributes.style "text-align" "left", Html.Attributes.style "padding" "12px" ] [ Html.text "Expires" ]
                    , Html.th [ Html.Attributes.style "text-align" "left", Html.Attributes.style "padding" "12px" ] [ Html.text "Created" ]
                    , Html.th [ Html.Attributes.style "text-align" "left", Html.Attributes.style "padding" "12px" ] [ Html.text "Actions" ]
                    ]
                ]
            , Html.tbody []
                (List.map viewTokenRow tokens)
            ]


viewTokenRow : APIToken -> Html Msg
viewTokenRow token =
    Html.tr [ Html.Attributes.style "border-bottom" "1px solid #ddd" ]
        [ Html.td [ Html.Attributes.style "padding" "12px" ] [ Html.text token.name ]
        , Html.td [ Html.Attributes.style "padding" "12px" ]
            [ Html.span
                [ Html.Attributes.style "background"
                    (if token.scope == "read_write" then
                        "#fff3e0"

                     else
                        "#e3f2fd"
                    )
                , Html.Attributes.style "padding" "2px 8px"
                , Html.Attributes.style "border-radius" "4px"
                , Html.Attributes.style "font-size" "12px"
                ]
                [ Html.text token.scope ]
            ]
        , Html.td [ Html.Attributes.style "padding" "12px" ] [ Html.text (formatDateString token.expiresAt) ]
        , Html.td [ Html.Attributes.style "padding" "12px" ] [ Html.text (formatDateString token.createdAt) ]
        , Html.td [ Html.Attributes.style "padding" "12px" ]
            [ Html.button
                [ Html.Events.onClick (OpenDeleteDialog token)
                , Html.Attributes.class "mdc-button"
                , Html.Attributes.style "color" "#f44336"
                ]
                [ Html.text "Revoke" ]
            ]
        ]


viewCreateDialog : Maybe CreateDialogState -> Html Msg
viewCreateDialog maybeDialog =
    case maybeDialog of
        Just dialog ->
            Components.viewDialogOverlay "Create API Token"
                [ Components.viewFormField "Token Name"
                    [ Html.Attributes.type_ "text"
                    , Html.Attributes.value dialog.name
                    , Html.Events.onInput SetCreateName
                    , Html.Attributes.placeholder "e.g., CI/CD Pipeline"
                    , Html.Attributes.required True
                    ]
                , Components.viewSelectField "Scope"
                    [ Html.Events.onInput SetCreateScope
                    , Html.Attributes.value dialog.scope
                    ]
                    [ Html.option [ Html.Attributes.value "read_only", Html.Attributes.selected (dialog.scope == "read_only") ] [ Html.text "Read Only" ]
                    , Html.option [ Html.Attributes.value "read_write", Html.Attributes.selected (dialog.scope == "read_write") ] [ Html.text "Read & Write" ]
                    ]
                , Components.viewSelectField "Expiry"
                    [ Html.Events.onInput SetCreateExpiry
                    , Html.Attributes.value (String.fromInt dialog.expiresInHours)
                    ]
                    [ Html.option [ Html.Attributes.value "168", Html.Attributes.selected (dialog.expiresInHours == 168) ] [ Html.text "1 Week" ]
                    , Html.option [ Html.Attributes.value "720", Html.Attributes.selected (dialog.expiresInHours == 720) ] [ Html.text "1 Month" ]
                    , Html.option [ Html.Attributes.value "2160", Html.Attributes.selected (dialog.expiresInHours == 2160) ] [ Html.text "3 Months" ]
                    , Html.option [ Html.Attributes.value "4320", Html.Attributes.selected (dialog.expiresInHours == 4320) ] [ Html.text "6 Months" ]
                    , Html.option [ Html.Attributes.value "8760", Html.Attributes.selected (dialog.expiresInHours == 8760) ] [ Html.text "1 Year" ]
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


viewDeleteDialog : APIToken -> Html Msg
viewDeleteDialog token =
    Components.viewDialogOverlay "Revoke Token"
        [ Html.div []
            [ Html.text ("Are you sure you want to revoke the token \"" ++ token.name ++ "\"? This action cannot be undone.")
            ]
        ]
        [ Html.button
            [ Html.Events.onClick CloseDeleteDialog
            , Html.Attributes.class "mdc-button"
            ]
            [ Html.text "Cancel" ]
        , Html.button
            [ Html.Events.onClick (ConfirmDelete token.id)
            , Html.Attributes.class "mdc-button mdc-button--raised"
            , Html.Attributes.style "background-color" "#f44336"
            ]
            [ Html.text "Revoke" ]
        ]



-- Integrations Views


viewIntegrationsSection : Model -> Html Msg
viewIntegrationsSection model =
    Html.div []
        [ Html.h2
            [ Html.Attributes.class "mdc-typography--headline5"
            , Html.Attributes.style "margin-bottom" "16px"
            ]
            [ Html.text "Integrations" ]
        , Html.div
            [ Html.Attributes.style "margin-bottom" "20px" ]
            [ Html.button
                [ Html.Events.onClick OpenCreateIntegrationDialog
                , Html.Attributes.class "mdc-button mdc-button--raised"
                ]
                [ Html.text "Connect Integration" ]
            ]
        , case model.testConnectionResult of
            Just (Ok message) ->
                Html.div
                    [ Html.Attributes.style "color" "#388e3c"
                    , Html.Attributes.style "background" "#e8f5e9"
                    , Html.Attributes.style "padding" "12px"
                    , Html.Attributes.style "border-radius" "4px"
                    , Html.Attributes.style "margin-bottom" "16px"
                    ]
                    [ Html.text message ]

            Just (Err message) ->
                Html.div
                    [ Html.Attributes.style "color" "#d32f2f"
                    , Html.Attributes.style "background" "#ffebee"
                    , Html.Attributes.style "padding" "12px"
                    , Html.Attributes.style "border-radius" "4px"
                    , Html.Attributes.style "margin-bottom" "16px"
                    ]
                    [ Html.text ("Connection failed: " ++ message) ]

            Nothing ->
                Html.text ""
        , if model.integrationsLoading && List.isEmpty model.integrations then
            Html.div [] [ Html.text "Loading..." ]

          else
            viewIntegrationsTable model.integrations
        ]


viewIntegrationsTable : List Integration -> Html Msg
viewIntegrationsTable integrations =
    if List.isEmpty integrations then
        Html.div
            [ Html.Attributes.style "color" "#666"
            , Html.Attributes.style "padding" "20px"
            ]
            [ Html.text "No integrations configured. Connect one to get started." ]

    else
        Html.table
            [ Html.Attributes.class "mdc-data-table__table"
            , Html.Attributes.style "width" "100%"
            , Html.Attributes.style "border-collapse" "collapse"
            ]
            [ Html.thead []
                [ Html.tr []
                    [ Html.th [ Html.Attributes.style "text-align" "left", Html.Attributes.style "padding" "12px" ] [ Html.text "Name" ]
                    , Html.th [ Html.Attributes.style "text-align" "left", Html.Attributes.style "padding" "12px" ] [ Html.text "Provider" ]
                    , Html.th [ Html.Attributes.style "text-align" "left", Html.Attributes.style "padding" "12px" ] [ Html.text "Status" ]
                    , Html.th [ Html.Attributes.style "text-align" "left", Html.Attributes.style "padding" "12px" ] [ Html.text "Actions" ]
                    ]
                ]
            , Html.tbody []
                (List.map viewIntegrationRow integrations)
            ]


viewIntegrationRow : Integration -> Html Msg
viewIntegrationRow integration =
    Html.tr [ Html.Attributes.style "border-bottom" "1px solid #ddd" ]
        [ Html.td [ Html.Attributes.style "padding" "12px" ] [ Html.text integration.name ]
        , Html.td [ Html.Attributes.style "padding" "12px" ]
            [ Html.span
                [ Html.Attributes.style "background"
                    (if integration.provider == "jira" then
                        "#e3f2fd"

                     else
                        "#f3e5f5"
                    )
                , Html.Attributes.style "padding" "2px 8px"
                , Html.Attributes.style "border-radius" "4px"
                , Html.Attributes.style "font-size" "12px"
                , Html.Attributes.style "text-transform" "capitalize"
                ]
                [ Html.text integration.provider ]
            ]
        , Html.td [ Html.Attributes.style "padding" "12px" ]
            [ Html.span
                [ Html.Attributes.style "color"
                    (if integration.isActive then
                        "#388e3c"

                     else
                        "#d32f2f"
                    )
                , Html.Attributes.style "font-weight" "bold"
                ]
                [ Html.text
                    (if integration.isActive then
                        "Active"

                     else
                        "Inactive"
                    )
                ]
            ]
        , Html.td [ Html.Attributes.style "padding" "12px" ]
            [ Html.button
                [ Html.Events.onClick (TestConnection integration.id)
                , Html.Attributes.class "mdc-button"
                , Html.Attributes.style "color" "#1976d2"
                ]
                [ Html.text "Test Connection" ]
            , Html.button
                [ Html.Events.onClick (OpenDeleteIntegrationDialog integration)
                , Html.Attributes.class "mdc-button"
                , Html.Attributes.style "color" "#f44336"
                ]
                [ Html.text "Delete" ]
            ]
        ]


viewCreateIntegrationDialog : Maybe CreateIntegrationDialogState -> Html Msg
viewCreateIntegrationDialog maybeDialog =
    case maybeDialog of
        Just dialog ->
            Components.viewDialogOverlay "Connect Integration"
                [ Components.viewFormField "Integration Name"
                    [ Html.Attributes.type_ "text"
                    , Html.Attributes.value dialog.name
                    , Html.Events.onInput SetIntegrationName
                    , Html.Attributes.placeholder "e.g., My Jira Instance"
                    , Html.Attributes.required True
                    ]
                , Components.viewSelectField "Provider"
                    [ Html.Events.onInput SetIntegrationProvider
                    , Html.Attributes.value dialog.provider
                    ]
                    [ Html.option [ Html.Attributes.value "jira", Html.Attributes.selected (dialog.provider == "jira") ] [ Html.text "Jira" ]
                    , Html.option [ Html.Attributes.value "github", Html.Attributes.selected (dialog.provider == "github") ] [ Html.text "GitHub" ]
                    ]
                , if dialog.provider == "jira" then
                    Html.div []
                        [ Components.viewFormField "Instance URL"
                            [ Html.Attributes.type_ "text"
                            , Html.Attributes.value dialog.jiraUrl
                            , Html.Events.onInput SetJiraUrl
                            , Html.Attributes.placeholder "https://your-domain.atlassian.net"
                            , Html.Attributes.required True
                            ]
                        , Components.viewFormField "Email"
                            [ Html.Attributes.type_ "email"
                            , Html.Attributes.value dialog.jiraEmail
                            , Html.Events.onInput SetJiraEmail
                            , Html.Attributes.placeholder "your-email@example.com"
                            , Html.Attributes.required True
                            ]
                        , Components.viewFormField "API Token"
                            [ Html.Attributes.type_ "password"
                            , Html.Attributes.value dialog.jiraApiToken
                            , Html.Events.onInput SetJiraApiToken
                            , Html.Attributes.placeholder "Your Jira API token"
                            , Html.Attributes.required True
                            ]
                        , Components.viewFormField "Default Project Key"
                            [ Html.Attributes.type_ "text"
                            , Html.Attributes.value dialog.jiraDefaultProject
                            , Html.Events.onInput SetJiraDefaultProject
                            , Html.Attributes.placeholder "e.g., PROJ"
                            ]
                        ]

                  else
                    Html.div []
                        [ Components.viewFormField "Personal Access Token"
                            [ Html.Attributes.type_ "password"
                            , Html.Attributes.value dialog.githubToken
                            , Html.Events.onInput SetGithubToken
                            , Html.Attributes.placeholder "ghp_..."
                            , Html.Attributes.required True
                            ]
                        , Components.viewFormField "Default Owner"
                            [ Html.Attributes.type_ "text"
                            , Html.Attributes.value dialog.githubDefaultOwner
                            , Html.Events.onInput SetGithubDefaultOwner
                            , Html.Attributes.placeholder "e.g., my-org"
                            ]
                        , Components.viewFormField "Default Repository"
                            [ Html.Attributes.type_ "text"
                            , Html.Attributes.value dialog.githubDefaultRepo
                            , Html.Events.onInput SetGithubDefaultRepo
                            , Html.Attributes.placeholder "e.g., my-repo"
                            ]
                        ]
                ]
                [ Html.button
                    [ Html.Events.onClick CloseCreateIntegrationDialog
                    , Html.Attributes.class "mdc-button"
                    ]
                    [ Html.text "Cancel" ]
                , Html.button
                    [ Html.Events.onClick SubmitCreateIntegration
                    , Html.Attributes.class "mdc-button mdc-button--raised"
                    ]
                    [ Html.text "Connect" ]
                ]

        Nothing ->
            Html.text ""


viewDeleteIntegrationDialog : Integration -> Html Msg
viewDeleteIntegrationDialog integration =
    Components.viewDialogOverlay "Delete Integration"
        [ Html.div []
            [ Html.text ("Are you sure you want to delete the integration \"" ++ integration.name ++ "\"? This will also remove all associated issue links. This action cannot be undone.")
            ]
        ]
        [ Html.button
            [ Html.Events.onClick CloseDeleteIntegrationDialog
            , Html.Attributes.class "mdc-button"
            ]
            [ Html.text "Cancel" ]
        , Html.button
            [ Html.Events.onClick (ConfirmDeleteIntegration integration.id)
            , Html.Attributes.class "mdc-button mdc-button--raised"
            , Html.Attributes.style "background-color" "#f44336"
            ]
            [ Html.text "Delete" ]
        ]



-- HELPERS


formatDateString : String -> String
formatDateString dateStr =
    String.left 10 dateStr


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
