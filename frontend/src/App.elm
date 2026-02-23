module App exposing (main)

import API
import Browser
import Browser.Navigation as Nav
import Html exposing (Html)
import Html.Attributes
import Html.Events
import Http
import Pages.Login as Login
import Pages.ProcedureDetail as ProcedureDetail
import Pages.Projects as Projects
import Pages.Register as Register
import Pages.TestProcedures as TestProcedures
import Pages.TestRunDetail as TestRunDetail
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
    , sessionCheckStatus : SessionCheckStatus
    , drawerOpen : Bool
    , loginModel : Login.Model
    , registerModel : Register.Model
    , projectsModel : Maybe Projects.Model
    , testProceduresModel : Maybe TestProcedures.Model
    , procedureDetailModel : Maybe ProcedureDetail.Model
    , testRunsModel : Maybe TestRuns.Model
    , testRunDetailModel : Maybe TestRunDetail.Model
    }


type Route
    = Login
    | Register
    | Projects
    | TestProcedures String
    | ProcedureDetail String String
    | TestRuns String
    | TestRunDetail String
    | NotFound


type SessionCheckStatus
    = CheckingSession
    | SessionChecked


init : () -> Url.Url -> Nav.Key -> ( Model, Cmd Msg )
init _ url key =
    let
        route =
            parseUrl url
    in
    ( { key = key
      , url = url
      , route = route
      , user = Nothing
      , sessionCheckStatus = CheckingSession
      , drawerOpen = False
      , loginModel = Login.init
      , registerModel = Register.init
      , projectsModel = Nothing
      , testProceduresModel = Nothing
      , procedureDetailModel = Nothing
      , testRunsModel = Nothing
      , testRunDetailModel = Nothing
      }
    , API.getMe SessionCheckResponse
    )



-- UPDATE


type Msg
    = LinkClicked Browser.UrlRequest
    | UrlChanged Url.Url
    | ToggleDrawer
    | CloseDrawer
    | SessionCheckResponse (Result Http.Error User)
    | LoginMsg Login.Msg
    | RegisterMsg Register.Msg
    | ProjectsMsg Projects.Msg
    | TestProceduresMsg TestProcedures.Msg
    | ProcedureDetailMsg ProcedureDetail.Msg
    | TestRunsMsg TestRuns.Msg
    | TestRunDetailMsg TestRunDetail.Msg
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
                        Register ->
                            ( { model | registerModel = Register.init }
                            , Cmd.none
                            )

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

                        ProcedureDetail projectId procedureId ->
                            let
                                ( pm, pc ) =
                                    ProcedureDetail.init projectId procedureId
                            in
                            ( { model | procedureDetailModel = Just pm }
                            , Cmd.map ProcedureDetailMsg pc
                            )

                        TestRuns procedureId ->
                            let
                                ( pm, pc ) =
                                    TestRuns.init procedureId
                            in
                            ( { model | testRunsModel = Just pm }
                            , Cmd.map TestRunsMsg pc
                            )

                        TestRunDetail runId ->
                            let
                                ( pm, pc ) =
                                    TestRunDetail.init runId
                            in
                            ( { model | testRunDetailModel = Just pm }
                            , Cmd.map TestRunDetailMsg pc
                            )

                        _ ->
                            ( model, Cmd.none )
            in
            ( { newModel | url = url, route = route }, cmd )

        SessionCheckResponse (Ok user) ->
            -- Valid session found - set user and navigate appropriately
            let
                ( newModel, cmd ) =
                    case model.route of
                        Login ->
                            -- User has valid session but on login page, redirect to projects
                            let
                                ( pm, pc ) =
                                    Projects.init
                            in
                            ( { model
                                | user = Just user
                                , route = Projects
                                , projectsModel = Just pm
                              }
                            , Cmd.batch
                                [ Nav.pushUrl model.key "/projects"
                                , Cmd.map ProjectsMsg pc
                                ]
                            )

                        Register ->
                            -- User has valid session but on register page, redirect to projects
                            let
                                ( pm, pc ) =
                                    Projects.init
                            in
                            ( { model
                                | user = Just user
                                , route = Projects
                                , projectsModel = Just pm
                              }
                            , Cmd.batch
                                [ Nav.pushUrl model.key "/projects"
                                , Cmd.map ProjectsMsg pc
                                ]
                            )

                        Projects ->
                            let
                                ( pm, pc ) =
                                    Projects.init
                            in
                            ( { model
                                | user = Just user
                                , projectsModel = Just pm
                              }
                            , Cmd.map ProjectsMsg pc
                            )

                        TestProcedures projectId ->
                            let
                                ( pm, pc ) =
                                    TestProcedures.init projectId
                            in
                            ( { model
                                | user = Just user
                                , testProceduresModel = Just pm
                              }
                            , Cmd.map TestProceduresMsg pc
                            )

                        ProcedureDetail projectId procedureId ->
                            let
                                ( pm, pc ) =
                                    ProcedureDetail.init projectId procedureId
                            in
                            ( { model
                                | user = Just user
                                , procedureDetailModel = Just pm
                              }
                            , Cmd.map ProcedureDetailMsg pc
                            )

                        TestRuns procedureId ->
                            let
                                ( pm, pc ) =
                                    TestRuns.init procedureId
                            in
                            ( { model
                                | user = Just user
                                , testRunsModel = Just pm
                              }
                            , Cmd.map TestRunsMsg pc
                            )

                        TestRunDetail runId ->
                            let
                                ( pm, pc ) =
                                    TestRunDetail.init runId
                            in
                            ( { model
                                | user = Just user
                                , testRunDetailModel = Just pm
                              }
                            , Cmd.map TestRunDetailMsg pc
                            )

                        NotFound ->
                            ( { model | user = Just user }, Cmd.none )
            in
            ( { newModel | sessionCheckStatus = SessionChecked }, cmd )

        SessionCheckResponse (Err _) ->
            -- No valid session or error - user stays on login/current page
            ( { model | sessionCheckStatus = SessionChecked }, Cmd.none )

        ToggleDrawer ->
            ( { model | drawerOpen = not model.drawerOpen }, Cmd.none )

        CloseDrawer ->
            ( { model | drawerOpen = False }, Cmd.none )

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

        RegisterMsg subMsg ->
            let
                ( newRegisterModel, cmd ) =
                    Register.update subMsg model.registerModel
            in
            case newRegisterModel.successfulUser of
                Just user ->
                    ( { model
                        | registerModel = { newRegisterModel | successfulUser = Nothing }
                        , user = Just user
                        , route = Projects
                      }
                    , Cmd.batch
                        [ Cmd.map RegisterMsg cmd
                        , Nav.pushUrl model.key "/projects"
                        ]
                    )

                Nothing ->
                    ( { model | registerModel = newRegisterModel }, Cmd.map RegisterMsg cmd )

        ProjectsMsg subMsg ->
            case model.projectsModel of
                Just projectsModel ->
                    let
                        ( newProjectsModel, cmd ) =
                            Projects.update subMsg projectsModel
                    in
                    case newProjectsModel.navigationTarget of
                        Just projectId ->
                            ( { model | projectsModel = Just { newProjectsModel | navigationTarget = Nothing } }
                            , Cmd.batch
                                [ Cmd.map ProjectsMsg cmd
                                , Nav.pushUrl model.key ("/projects/" ++ projectId ++ "/procedures")
                                ]
                            )

                        Nothing ->
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
                    case newModel.navigationTarget of
                        Just procedureId ->
                            ( { model | testProceduresModel = Just { newModel | navigationTarget = Nothing } }
                            , Cmd.batch
                                [ Cmd.map TestProceduresMsg cmd
                                , Nav.pushUrl model.key
                                    ("/projects/" ++ newModel.projectId ++ "/procedures/" ++ procedureId)
                                ]
                            )

                        Nothing ->
                            ( { model | testProceduresModel = Just newModel }
                            , Cmd.map TestProceduresMsg cmd
                            )

                Nothing ->
                    ( model, Cmd.none )

        ProcedureDetailMsg subMsg ->
            case model.procedureDetailModel of
                Just procedureDetailModel ->
                    let
                        ( newModel, cmd ) =
                            ProcedureDetail.update subMsg procedureDetailModel
                    in
                    ( { model | procedureDetailModel = Just newModel }
                    , Cmd.map ProcedureDetailMsg cmd
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

        TestRunDetailMsg subMsg ->
            case model.testRunDetailModel of
                Just testRunDetailModel ->
                    let
                        ( newModel, cmd ) =
                            TestRunDetail.update subMsg testRunDetailModel
                    in
                    ( { model | testRunDetailModel = Just newModel }
                    , Cmd.map TestRunDetailMsg cmd
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
                , sessionCheckStatus = SessionChecked
              }
            , Nav.pushUrl model.key "/"
            )

        LogoutResponse (Err _) ->
            -- Even if logout fails on server, clear local state
            ( { model
                | user = Nothing
                , route = Login
                , drawerOpen = False
                , sessionCheckStatus = SessionChecked
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

                /* Responsive main content padding */
                .app-main-content {
                    padding: 16px;
                }

                @media (min-width: 600px) {
                    .app-main-content {
                        padding: 24px;
                    }
                }

                @media (min-width: 1024px) {
                    .app-main-content {
                        padding: 32px 48px;
                    }
                }

                /* Responsive page header spacing */
                .page-header {
                    display: flex;
                    justify-content: space-between;
                    align-items: center;
                    margin-bottom: 20px;
                    margin-top: 0;
                }

                @media (min-width: 600px) {
                    .page-header {
                        margin-top: 8px;
                    }
                }

                @media (min-width: 1024px) {
                    .page-header {
                        margin-top: 16px;
                    }
                }
            """ ]
        , viewTopAppBar model
        , if model.drawerOpen && model.user /= Nothing then
            Html.div
                [ Html.Events.onClick CloseDrawer
                , Html.Attributes.style "position" "fixed"
                , Html.Attributes.style "top" "0"
                , Html.Attributes.style "left" "0"
                , Html.Attributes.style "width" "100%"
                , Html.Attributes.style "height" "100%"
                , Html.Attributes.style "z-index" "50"
                ]
                []

          else
            Html.text ""
        , if model.user /= Nothing then
            viewDrawer model

          else
            Html.text ""
        , Html.main_
            [ Html.Attributes.class "app-main-content" ]
            [ viewContent model ]
        ]
    }


viewTopAppBar : Model -> Html Msg
viewTopAppBar model =
    Html.header
        [ Html.Attributes.style "background" "#6200ee"
        , Html.Attributes.style "color" "white"
        , Html.Attributes.style "padding" "16px 24px"
        , Html.Attributes.style "display" "flex"
        , Html.Attributes.style "flex-direction" "row"
        , Html.Attributes.style "justify-content" "space-between"
        , Html.Attributes.style "align-items" "center"
        , Html.Attributes.style "box-shadow" "0 2px 4px rgba(0,0,0,0.1)"
        , Html.Attributes.style "position" "sticky"
        , Html.Attributes.style "top" "0"
        , Html.Attributes.style "z-index" "100"
        , Html.Attributes.style "box-sizing" "border-box"
        ]
        [ Html.div
            [ Html.Attributes.style "display" "flex"
            , Html.Attributes.style "align-items" "center"
            , Html.Attributes.style "gap" "16px"
            ]
            [ if model.user /= Nothing then
                Html.button
                    [ Html.Events.onClick ToggleDrawer
                    , Html.Attributes.style "color" "white"
                    , Html.Attributes.style "background" "none"
                    , Html.Attributes.style "border" "none"
                    , Html.Attributes.style "cursor" "pointer"
                    , Html.Attributes.style "font-size" "24px"
                    , Html.Attributes.style "padding" "8px"
                    , Html.Attributes.style "display" "flex"
                    , Html.Attributes.style "align-items" "center"
                    ]
                    [ Html.text "☰" ]

              else
                Html.text ""
            , Html.h1
                [ Html.Attributes.style "margin" "0"
                , Html.Attributes.style "font-size" "20px"
                , Html.Attributes.style "font-weight" "500"
                ]
                [ Html.text "UI Automation" ]
            ]
        , case model.user of
            Just user ->
                Html.div
                    [ Html.Attributes.style "display" "flex"
                    , Html.Attributes.style "align-items" "center"
                    , Html.Attributes.style "gap" "16px"
                    , Html.Attributes.style "margin-right" "8px"
                    ]
                    [ Html.span
                        [ Html.Attributes.style "font-size" "14px"
                        , Html.Attributes.style "white-space" "nowrap"
                        ]
                        [ Html.text user.username ]
                    , Html.button
                        [ Html.Events.onClick Logout
                        , Html.Attributes.style "color" "white"
                        , Html.Attributes.style "background" "rgba(255, 255, 255, 0.1)"
                        , Html.Attributes.style "border" "1px solid rgba(255, 255, 255, 0.3)"
                        , Html.Attributes.style "border-radius" "4px"
                        , Html.Attributes.style "cursor" "pointer"
                        , Html.Attributes.style "font-size" "14px"
                        , Html.Attributes.style "padding" "6px 12px"
                        , Html.Attributes.style "display" "flex"
                        , Html.Attributes.style "align-items" "center"
                        ]
                        [ Html.text "Logout" ]
                    ]

            Nothing ->
                Html.text ""
        ]


viewDrawer : Model -> Html Msg
viewDrawer model =
    if model.drawerOpen then
        Html.div
            [ Html.Attributes.style "position" "fixed"
            , Html.Attributes.style "top" "64px"
            , Html.Attributes.style "left" "0"
            , Html.Attributes.style "width" "256px"
            , Html.Attributes.style "height" "calc(100% - 64px)"
            , Html.Attributes.style "z-index" "60"
            , Html.Attributes.style "background" "#fff"
            , Html.Attributes.style "box-shadow" "2px 0 8px rgba(0,0,0,0.2)"
            , Html.Attributes.style "overflow-y" "auto"
            ]
            [ Html.nav [ Html.Attributes.class "mdc-list", Html.Attributes.style "padding-top" "24px" ]
                [ Html.a
                    [ Html.Attributes.href "/projects"
                    , Html.Attributes.class "mdc-list-item"
                    ]
                    [ Html.text "Projects" ]
                ]
            ]

    else
        Html.text ""


viewContent : Model -> Html Msg
viewContent model =
    case model.sessionCheckStatus of
        CheckingSession ->
            -- Show loading while checking session
            Html.div
                [ Html.Attributes.style "display" "flex"
                , Html.Attributes.style "justify-content" "center"
                , Html.Attributes.style "align-items" "center"
                , Html.Attributes.style "height" "80vh"
                ]
                [ Html.div []
                    [ Html.text "Loading..." ]
                ]

        SessionChecked ->
            -- Show content based on route and authentication status
            case model.route of
                Login ->
                    Html.map LoginMsg (Login.view model.loginModel)

                Register ->
                    Html.map RegisterMsg (Register.view model.registerModel)

                Projects ->
                    case model.user of
                        Just _ ->
                            case model.projectsModel of
                                Just projectsModel ->
                                    Html.map ProjectsMsg (Projects.view projectsModel)

                                Nothing ->
                                    Html.div [] [ Html.text "Loading..." ]

                        Nothing ->
                            Html.map LoginMsg (Login.view model.loginModel)

                TestProcedures _ ->
                    case model.user of
                        Just _ ->
                            case model.testProceduresModel of
                                Just testProceduresModel ->
                                    Html.map TestProceduresMsg (TestProcedures.view testProceduresModel)

                                Nothing ->
                                    Html.div [] [ Html.text "Loading..." ]

                        Nothing ->
                            Html.map LoginMsg (Login.view model.loginModel)

                ProcedureDetail _ _ ->
                    case model.user of
                        Just _ ->
                            case model.procedureDetailModel of
                                Just procedureDetailModel ->
                                    Html.map ProcedureDetailMsg (ProcedureDetail.view procedureDetailModel)

                                Nothing ->
                                    Html.div [] [ Html.text "Loading..." ]

                        Nothing ->
                            Html.map LoginMsg (Login.view model.loginModel)

                TestRuns _ ->
                    case model.user of
                        Just _ ->
                            case model.testRunsModel of
                                Just testRunsModel ->
                                    Html.map TestRunsMsg (TestRuns.view testRunsModel)

                                Nothing ->
                                    Html.div [] [ Html.text "Loading..." ]

                        Nothing ->
                            Html.map LoginMsg (Login.view model.loginModel)

                TestRunDetail _ ->
                    case model.user of
                        Just _ ->
                            case model.testRunDetailModel of
                                Just testRunDetailModel ->
                                    Html.map TestRunDetailMsg (TestRunDetail.view testRunDetailModel)

                                Nothing ->
                                    Html.div [] [ Html.text "Loading..." ]

                        Nothing ->
                            Html.map LoginMsg (Login.view model.loginModel)

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
        , Parser.map Register (Parser.s "register")
        , Parser.map Projects (Parser.s "projects")
        , Parser.map ProcedureDetail (Parser.s "projects" </> Parser.string </> Parser.s "procedures" </> Parser.string)
        , Parser.map TestProcedures (Parser.s "projects" </> Parser.string </> Parser.s "procedures")
        , Parser.map TestRuns (Parser.s "procedures" </> Parser.string </> Parser.s "runs")
        , Parser.map TestRunDetail (Parser.s "runs" </> Parser.string)
        ]
