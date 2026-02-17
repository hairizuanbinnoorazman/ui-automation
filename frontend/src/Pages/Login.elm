module Pages.Login exposing (Model, Msg, init, update, view)

import API
import Html exposing (Html)
import Html.Attributes
import Html.Events
import Http
import Types exposing (LoginCredentials, RegisterCredentials, User)



-- MODEL


type alias Model =
    { mode : Mode
    , email : String
    , username : String
    , password : String
    , confirmPassword : String
    , error : Maybe String
    , loading : Bool
    }


type Mode
    = LoginMode
    | RegisterMode


init : Model
init =
    { mode = LoginMode
    , email = ""
    , username = ""
    , password = ""
    , confirmPassword = ""
    , error = Nothing
    , loading = False
    }



-- UPDATE


type Msg
    = SetMode Mode
    | SetEmail String
    | SetUsername String
    | SetPassword String
    | SetConfirmPassword String
    | SubmitLogin
    | SubmitRegister
    | LoginResponse (Result Http.Error User)
    | RegisterResponse (Result Http.Error User)


update : Msg -> Model -> ( Model, Cmd Msg )
update msg model =
    case msg of
        SetMode mode ->
            ( { model
                | mode = mode
                , error = Nothing
                , email = ""
                , username = ""
                , password = ""
                , confirmPassword = ""
              }
            , Cmd.none
            )

        SetEmail email ->
            ( { model | email = email }, Cmd.none )

        SetUsername username ->
            ( { model | username = username }, Cmd.none )

        SetPassword password ->
            ( { model | password = password }, Cmd.none )

        SetConfirmPassword confirmPassword ->
            ( { model | confirmPassword = confirmPassword }, Cmd.none )

        SubmitLogin ->
            if String.isEmpty model.email || String.isEmpty model.password then
                ( { model | error = Just "Email and password are required" }, Cmd.none )

            else
                ( { model | loading = True, error = Nothing }
                , API.login
                    { email = model.email
                    , password = model.password
                    }
                    LoginResponse
                )

        SubmitRegister ->
            if String.isEmpty model.email || String.isEmpty model.username || String.isEmpty model.password then
                ( { model | error = Just "All fields are required" }, Cmd.none )

            else if model.password /= model.confirmPassword then
                ( { model | error = Just "Passwords do not match" }, Cmd.none )

            else if String.length model.password < 8 then
                ( { model | error = Just "Password must be at least 8 characters" }, Cmd.none )

            else
                ( { model | loading = True, error = Nothing }
                , API.register
                    { email = model.email
                    , username = model.username
                    , password = model.password
                    }
                    RegisterResponse
                )

        LoginResponse (Ok user) ->
            ( { model | loading = False }
            , Cmd.none
            )

        LoginResponse (Err error) ->
            ( { model
                | loading = False
                , error = Just (httpErrorToString error)
              }
            , Cmd.none
            )

        RegisterResponse (Ok user) ->
            ( { model | loading = False }
            , Cmd.none
            )

        RegisterResponse (Err error) ->
            ( { model
                | loading = False
                , error = Just (httpErrorToString error)
              }
            , Cmd.none
            )



-- VIEW


view : Model -> Html Msg
view model =
    Html.div
        [ Html.Attributes.style "display" "flex"
        , Html.Attributes.style "justify-content" "center"
        , Html.Attributes.style "align-items" "center"
        , Html.Attributes.style "min-height" "100vh"
        ]
        [ Html.div
            [ Html.Attributes.class "mdc-card"
            , Html.Attributes.style "padding" "24px"
            , Html.Attributes.style "max-width" "400px"
            ]
            [ Html.h2
                [ Html.Attributes.class "mdc-typography--headline4" ]
                [ Html.text
                    (if model.mode == LoginMode then
                        "Login"

                     else
                        "Register"
                    )
                ]
            , case model.error of
                Just err ->
                    Html.div
                        [ Html.Attributes.class "mdc-typography--body1"
                        , Html.Attributes.style "color" "red"
                        , Html.Attributes.style "margin-bottom" "16px"
                        ]
                        [ Html.text err ]

                Nothing ->
                    Html.text ""
            , if model.mode == LoginMode then
                viewLoginForm model

              else
                viewRegisterForm model
            , Html.div [ Html.Attributes.style "margin-top" "16px" ]
                [ Html.button
                    [ Html.Events.onClick
                        (SetMode
                            (if model.mode == LoginMode then
                                RegisterMode

                             else
                                LoginMode
                            )
                        )
                    , Html.Attributes.class "mdc-button"
                    ]
                    [ Html.text
                        (if model.mode == LoginMode then
                            "Need an account? Register"

                         else
                            "Have an account? Login"
                        )
                    ]
                ]
            ]
        ]


viewLoginForm : Model -> Html Msg
viewLoginForm model =
    Html.div []
        [ Html.div [ Html.Attributes.style "margin-bottom" "16px" ]
            [ Html.label [] [ Html.text "Email" ]
            , Html.input
                [ Html.Attributes.type_ "email"
                , Html.Attributes.value model.email
                , Html.Events.onInput SetEmail
                , Html.Attributes.required True
                , Html.Attributes.class "mdc-text-field__input"
                , Html.Attributes.style "width" "100%"
                ]
                []
            ]
        , Html.div [ Html.Attributes.style "margin-bottom" "16px" ]
            [ Html.label [] [ Html.text "Password" ]
            , Html.input
                [ Html.Attributes.type_ "password"
                , Html.Attributes.value model.password
                , Html.Events.onInput SetPassword
                , Html.Attributes.required True
                , Html.Attributes.class "mdc-text-field__input"
                , Html.Attributes.style "width" "100%"
                ]
                []
            ]
        , Html.button
            [ Html.Events.onClick SubmitLogin
            , Html.Attributes.disabled model.loading
            , Html.Attributes.class "mdc-button mdc-button--raised"
            ]
            [ Html.text
                (if model.loading then
                    "Logging in..."

                 else
                    "Login"
                )
            ]
        ]


viewRegisterForm : Model -> Html Msg
viewRegisterForm model =
    Html.div []
        [ Html.div [ Html.Attributes.style "margin-bottom" "16px" ]
            [ Html.label [] [ Html.text "Email" ]
            , Html.input
                [ Html.Attributes.type_ "email"
                , Html.Attributes.value model.email
                , Html.Events.onInput SetEmail
                , Html.Attributes.required True
                , Html.Attributes.class "mdc-text-field__input"
                , Html.Attributes.style "width" "100%"
                ]
                []
            ]
        , Html.div [ Html.Attributes.style "margin-bottom" "16px" ]
            [ Html.label [] [ Html.text "Username" ]
            , Html.input
                [ Html.Attributes.type_ "text"
                , Html.Attributes.value model.username
                , Html.Events.onInput SetUsername
                , Html.Attributes.required True
                , Html.Attributes.class "mdc-text-field__input"
                , Html.Attributes.style "width" "100%"
                ]
                []
            ]
        , Html.div [ Html.Attributes.style "margin-bottom" "16px" ]
            [ Html.label [] [ Html.text "Password" ]
            , Html.input
                [ Html.Attributes.type_ "password"
                , Html.Attributes.value model.password
                , Html.Events.onInput SetPassword
                , Html.Attributes.required True
                , Html.Attributes.class "mdc-text-field__input"
                , Html.Attributes.style "width" "100%"
                ]
                []
            ]
        , Html.div [ Html.Attributes.style "margin-bottom" "16px" ]
            [ Html.label [] [ Html.text "Confirm Password" ]
            , Html.input
                [ Html.Attributes.type_ "password"
                , Html.Attributes.value model.confirmPassword
                , Html.Events.onInput SetConfirmPassword
                , Html.Attributes.required True
                , Html.Attributes.class "mdc-text-field__input"
                , Html.Attributes.style "width" "100%"
                ]
                []
            ]
        , Html.button
            [ Html.Events.onClick SubmitRegister
            , Html.Attributes.disabled model.loading
            , Html.Attributes.class "mdc-button mdc-button--raised"
            ]
            [ Html.text
                (if model.loading then
                    "Registering..."

                 else
                    "Register"
                )
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
