module Pages.Login exposing (Model, Msg, init, update, view)

import API
import Html exposing (Html)
import Http
import Material.Button as Button
import Material.Card as Card
import Material.LayoutGrid as LayoutGrid
import Material.TextField as TextField
import Material.Typography as Typography
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
    LayoutGrid.layoutGrid []
        [ LayoutGrid.cell
            [ LayoutGrid.span4Desktop
            , LayoutGrid.span4Tablet
            , LayoutGrid.span4Phone
            , LayoutGrid.align LayoutGrid.Middle
            ]
            []
        , LayoutGrid.cell
            [ LayoutGrid.span4Desktop
            , LayoutGrid.span4Tablet
            , LayoutGrid.span4Phone
            , LayoutGrid.align LayoutGrid.Middle
            ]
            [ Card.card
                (Card.config |> Card.setAttributes [ Typography.typography ])
                { blocks =
                    [ Card.block <|
                        Html.div []
                            [ Html.h2
                                [ Typography.headline4 ]
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
                                        [ Typography.body1
                                        ]
                                        [ Html.text err ]

                                Nothing ->
                                    Html.text ""
                            , if model.mode == LoginMode then
                                viewLoginForm model

                              else
                                viewRegisterForm model
                            , Html.div []
                                [ Html.button
                                    []
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
                , actions = Nothing
                }
            ]
        , LayoutGrid.cell
            [ LayoutGrid.span4Desktop
            , LayoutGrid.span4Tablet
            , LayoutGrid.span4Phone
            , LayoutGrid.align LayoutGrid.Middle
            ]
            []
        ]


viewLoginForm : Model -> Html Msg
viewLoginForm model =
    Html.div []
        [ TextField.filled
            (TextField.config
                |> TextField.setLabel (Just "Email")
                |> TextField.setValue (Just model.email)
                |> TextField.setOnInput (Just SetEmail)
                |> TextField.setType (Just "email")
                |> TextField.setRequired True
            )
        , TextField.filled
            (TextField.config
                |> TextField.setLabel (Just "Password")
                |> TextField.setValue (Just model.password)
                |> TextField.setOnInput (Just SetPassword)
                |> TextField.setType (Just "password")
                |> TextField.setRequired True
            )
        , Button.raised
            (Button.config
                |> Button.setOnClick (Just SubmitLogin)
                |> Button.setDisabled model.loading
            )
            (if model.loading then
                "Logging in..."

             else
                "Login"
            )
        ]


viewRegisterForm : Model -> Html Msg
viewRegisterForm model =
    Html.div []
        [ TextField.filled
            (TextField.config
                |> TextField.setLabel (Just "Email")
                |> TextField.setValue (Just model.email)
                |> TextField.setOnInput (Just SetEmail)
                |> TextField.setType (Just "email")
                |> TextField.setRequired True
            )
        , TextField.filled
            (TextField.config
                |> TextField.setLabel (Just "Username")
                |> TextField.setValue (Just model.username)
                |> TextField.setOnInput (Just SetUsername)
                |> TextField.setRequired True
            )
        , TextField.filled
            (TextField.config
                |> TextField.setLabel (Just "Password")
                |> TextField.setValue (Just model.password)
                |> TextField.setOnInput (Just SetPassword)
                |> TextField.setType (Just "password")
                |> TextField.setRequired True
            )
        , TextField.filled
            (TextField.config
                |> TextField.setLabel (Just "Confirm Password")
                |> TextField.setValue (Just model.confirmPassword)
                |> TextField.setOnInput (Just SetConfirmPassword)
                |> TextField.setType (Just "password")
                |> TextField.setRequired True
            )
        , Button.raised
            (Button.config
                |> Button.setOnClick (Just SubmitRegister)
                |> Button.setDisabled model.loading
            )
            (if model.loading then
                "Registering..."

             else
                "Register"
            )
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
