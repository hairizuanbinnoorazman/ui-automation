module App exposing (main)

import API
import Browser
import Browser.Navigation as Nav
import Html exposing (Html)
import Html.Attributes
import Html.Events
import Http
import Pages.Login as Login
import Pages.Projects as Projects
import Pages.TestProcedures as TestProcedures
import Pages.TestRuns as TestRuns
import Types exposing (User)
import Url
import Url.Parser as Parser exposing (Parser, (</>))



-- MAIN


main : Program () Model Msg
main =
    Browser.application
        { init = init
        , view = view
        , update = update
        , subscriptions = subscriptions
        , onUrlChange = UrlChanged
        , onUrlRequest = LinkClicked
        }



-- MODEL


type alias Model =
    { key : Nav.Key
    , url : Url.Url
    , route : Route
    , user : Maybe User
    , drawerOpen : Bool
    , loginModel : Login.Model
    , projectsModel : Maybe Projects.Model
    , testProceduresModel : Maybe TestProcedures.Model
    , testRunsModel : Maybe TestRuns.Model
    }


type Route
    = Login
    | Projects
    | TestProcedures String
    | TestRuns String
    | NotFound


init : () -> Url.Url -> Nav.Key -> ( Model, Cmd Msg )
init _ url key =
    let
        route =
            parseUrl url

        ( loginModel, loginCmd ) =
            ( Login.init, Cmd.none )

        ( projectsModel, projectsCmd ) =
            case route of
                Projects ->
                    let
                        ( pm, pc ) =
                            Projects.init
                    in
                    ( Just pm, Cmd.map ProjectsMsg pc )

                _ ->
                    ( Nothing, Cmd.none )
    in
    ( { key = key
      , url = url
      , route = route
      , user = Nothing
      , drawerOpen = False
      , loginModel = loginModel
      , projectsModel = projectsModel
      , testProceduresModel = Nothing
      , testRunsModel = Nothing
      }
    , Cmd.batch
        [ loginCmd
        , projectsCmd
        ]
    )



-- UPDATE


type Msg
    = LinkClicked Browser.UrlRequest
    | UrlChanged Url.Url
    | ToggleDrawer
    | LoginMsg Login.Msg
    | ProjectsMsg Projects.Msg
    | TestProceduresMsg TestProcedures.Msg
    | TestRunsMsg TestRuns.Msg
    | Logout
    | LogoutResponse (Result Http.Error ())


update : Msg -> Model -> ( Model, Cmd Msg )
update msg model =
    case msg of
        LinkClicked urlRequest ->
            case urlRequest of
                Browser.Internal url ->
                    ( model, Nav.pushUrl model.key (Url.toString url) )

                Browser.External href ->
                    ( model, Nav.load href )

        UrlChanged url ->
            let
                route =
                    parseUrl url

                ( newModel, cmd ) =
                    case route of
                        Projects ->
                            case model.projectsModel of
                                Just _ ->
                                    ( model, Cmd.none )

                                Nothing ->
                                    let
                                        ( pm, pc ) =
                                            Projects.init
                                    in
                                    ( { model | projectsModel = Just pm }
                                    , Cmd.map ProjectsMsg pc
                                    )

                        TestProcedures projectId ->
                            let
                                ( pm, pc ) =
                                    TestProcedures.init projectId
                            in
                            ( { model | testProceduresModel = Just pm }
                            , Cmd.map TestProceduresMsg pc
                            )

                        TestRuns procedureId ->
                            let
                                ( pm, pc ) =
                                    TestRuns.init procedureId
                            in
                            ( { model | testRunsModel = Just pm }
                            , Cmd.map TestRunsMsg pc
                            )

                        _ ->
                            ( model, Cmd.none )
            in
            ( { newModel | url = url, route = route }, cmd )

        ToggleDrawer ->
            ( { model | drawerOpen = not model.drawerOpen }, Cmd.none )

        LoginMsg subMsg ->
            let
                ( newLoginModel, cmd ) =
                    Login.update subMsg model.loginModel
            in
            case newLoginModel.successfulUser of
                Just user ->
                    ( { model
                        | loginModel = { newLoginModel | successfulUser = Nothing }
                        , user = Just user
                        , route = Projects
                      }
                    , Cmd.batch
                        [ Cmd.map LoginMsg cmd
                        , Nav.pushUrl model.key "/projects"
                        ]
                    )

                Nothing ->
                    ( { model | loginModel = newLoginModel }, Cmd.map LoginMsg cmd )

        ProjectsMsg subMsg ->
            case model.projectsModel of
                Just projectsModel ->
                    let
                        ( newProjectsModel, cmd ) =
                            Projects.update subMsg projectsModel
                    in
                    ( { model | projectsModel = Just newProjectsModel }
                    , Cmd.map ProjectsMsg cmd
                    )

                Nothing ->
                    ( model, Cmd.none )

        TestProceduresMsg subMsg ->
            case model.testProceduresModel of
                Just testProceduresModel ->
                    let
                        ( newModel, cmd ) =
                            TestProcedures.update subMsg testProceduresModel
                    in
                    ( { model | testProceduresModel = Just newModel }
                    , Cmd.map TestProceduresMsg cmd
                    )

                Nothing ->
                    ( model, Cmd.none )

        TestRunsMsg subMsg ->
            case model.testRunsModel of
                Just testRunsModel ->
                    let
                        ( newModel, cmd ) =
                            TestRuns.update subMsg testRunsModel
                    in
                    ( { model | testRunsModel = Just newModel }
                    , Cmd.map TestRunsMsg cmd
                    )

                Nothing ->
                    ( model, Cmd.none )

        Logout ->
            ( model
            , API.logout LogoutResponse
            )

        LogoutResponse (Ok ()) ->
            ( { model
                | user = Nothing
                , route = Login
                , drawerOpen = False
              }
            , Nav.pushUrl model.key "/"
            )

        LogoutResponse (Err _) ->
            -- Even if logout fails on server, clear local state
            ( { model
                | user = Nothing
                , route = Login
                , drawerOpen = False
              }
            , Nav.pushUrl model.key "/"
            )



-- SUBSCRIPTIONS


subscriptions : Model -> Sub Msg
subscriptions _ =
    Sub.none



-- VIEW


view : Model -> Browser.Document Msg
view model =
    { title = "UI Automation"
    , body =
        [ Html.node "style"
            []
            [ Html.text """
                body {
                    padding: 0;
                    margin: 0;
                    font-family: Roboto, sans-serif;
                    -webkit-font-smoothing: antialiased;
                    -moz-osx-font-smoothing: grayscale;
                }
            """ ]
        , Html.div
            [ Html.Attributes.class "mdc-drawer-app-content" ]
            [ viewTopAppBar model
            , Html.div
                [ Html.Attributes.style "display" "flex" ]
                [ if model.user /= Nothing then
                    viewDrawer model

                  else
                    Html.text ""
                , Html.main_
                    [ Html.Attributes.style "flex-grow" "1"
                    , Html.Attributes.style "padding" "20px"
                    ]
                    [ viewContent model ]
                ]
            ]
        ]
    }


viewTopAppBar : Model -> Html Msg
viewTopAppBar model =
    Html.header
        [ Html.Attributes.class "mdc-top-app-bar"
        , Html.Attributes.style "background" "#6200ee"
        , Html.Attributes.style "color" "white"
        , Html.Attributes.style "padding" "16px"
        , Html.Attributes.style "display" "flex"
        , Html.Attributes.style "justify-content" "space-between"
        , Html.Attributes.style "align-items" "center"
        ]
        [ Html.div
            [ Html.Attributes.style "display" "flex"
            , Html.Attributes.style "align-items" "center"
            , Html.Attributes.style "gap" "16px"
            ]
            [ if model.user /= Nothing then
                Html.button
                    [ Html.Events.onClick ToggleDrawer
                    , Html.Attributes.class "mdc-icon-button"
                    , Html.Attributes.style "color" "white"
                    , Html.Attributes.style "background" "none"
                    , Html.Attributes.style "border" "none"
                    , Html.Attributes.style "cursor" "pointer"
                    ]
                    [ Html.text "☰" ]

              else
                Html.text ""
            , Html.h1
                [ Html.Attributes.class "mdc-top-app-bar__title"
                , Html.Attributes.style "margin" "0"
                , Html.Attributes.style "font-size" "20px"
                ]
                [ Html.text "UI Automation" ]
            ]
        , case model.user of
            Just user ->
                Html.div
                    [ Html.Attributes.style "display" "flex"
                    , Html.Attributes.style "align-items" "center"
                    , Html.Attributes.style "gap" "16px"
                    ]
                    [ Html.span [] [ Html.text user.username ]
                    , Html.button
                        [ Html.Events.onClick Logout
                        , Html.Attributes.class "mdc-icon-button"
                        , Html.Attributes.style "color" "white"
                        , Html.Attributes.style "background" "none"
                        , Html.Attributes.style "border" "none"
                        , Html.Attributes.style "cursor" "pointer"
                        ]
                        [ Html.text "⎋" ]
                    ]

            Nothing ->
                Html.text ""
        ]


viewDrawer : Model -> Html Msg
viewDrawer model =
    Html.div
        [ Html.Attributes.class "mdc-drawer mdc-drawer--dismissible"
        , Html.Attributes.classList [ ( "mdc-drawer--open", model.drawerOpen ) ]
        , Html.Attributes.style "width" "256px"
        ]
        [ Html.div [ Html.Attributes.class "mdc-drawer__content" ]
            [ Html.nav [ Html.Attributes.class "mdc-list" ]
                [ Html.a
                    [ Html.Attributes.href "/projects"
                    , Html.Attributes.class "mdc-list-item"
                    ]
                    [ Html.text "Projects" ]
                ]
            ]
        ]


viewContent : Model -> Html Msg
viewContent model =
    case model.route of
        Login ->
            Html.map LoginMsg (Login.view model.loginModel)

        Projects ->
            case model.projectsModel of
                Just projectsModel ->
                    Html.map ProjectsMsg (Projects.view projectsModel)

                Nothing ->
                    Html.div [] [ Html.text "Loading..." ]

        TestProcedures _ ->
            case model.testProceduresModel of
                Just testProceduresModel ->
                    Html.map TestProceduresMsg (TestProcedures.view testProceduresModel)

                Nothing ->
                    Html.div [] [ Html.text "Loading..." ]

        TestRuns _ ->
            case model.testRunsModel of
                Just testRunsModel ->
                    Html.map TestRunsMsg (TestRuns.view testRunsModel)

                Nothing ->
                    Html.div [] [ Html.text "Loading..." ]

        NotFound ->
            Html.div []
                [ Html.h1 [] [ Html.text "404 Not Found" ]
                , Html.p [] [ Html.text "The page you're looking for doesn't exist." ]
                ]



-- URL PARSING


parseUrl : Url.Url -> Route
parseUrl url =
    case Parser.parse routeParser url of
        Just route ->
            route

        Nothing ->
            Login


routeParser : Parser (Route -> a) a
routeParser =
    Parser.oneOf
        [ Parser.map Login Parser.top
        , Parser.map Projects (Parser.s "projects")
        , Parser.map TestProcedures (Parser.s "projects" </> Parser.string </> Parser.s "procedures")
        , Parser.map TestRuns (Parser.s "procedures" </> Parser.string </> Parser.s "runs")
        ]
