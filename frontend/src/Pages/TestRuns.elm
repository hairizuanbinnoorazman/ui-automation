module Pages.TestRuns exposing (Model, Msg, init, update, view)

import API
import Html exposing (Html)
import Html.Attributes
import Html.Events
import Http
import Time
import Types exposing (PaginatedResponse, TestRun, TestRunStatus)



-- MODEL


type alias Model =
    { procedureId : String
    , runs : List TestRun
    , total : Int
    , limit : Int
    , offset : Int
    , loading : Bool
    , error : Maybe String
    }


init : String -> ( Model, Cmd Msg )
init procedureId =
    ( { procedureId = procedureId
      , runs = []
      , total = 0
      , limit = 10
      , offset = 0
      , loading = True
      , error = Nothing
      }
    , API.getTestRuns procedureId 10 0 RunsResponse
    )



-- UPDATE


type Msg
    = RunsResponse (Result Http.Error (PaginatedResponse TestRun))
    | LoadPage Int
    | SubmitCreate
    | CreateResponse (Result Http.Error TestRun)


update : Msg -> Model -> ( Model, Cmd Msg )
update msg model =
    case msg of
        RunsResponse (Ok response) ->
            ( { model
                | runs = response.items
                , total = response.total
                , loading = False
                , error = Nothing
              }
            , Cmd.none
            )

        RunsResponse (Err error) ->
            ( { model
                | loading = False
                , error = Just (httpErrorToString error)
              }
            , Cmd.none
            )

        LoadPage offset ->
            ( { model | loading = True, offset = offset }
            , API.getTestRuns model.procedureId model.limit offset RunsResponse
            )

        SubmitCreate ->
            ( { model | loading = True }
            , API.createTestRun model.procedureId CreateResponse
            )

        CreateResponse (Ok _) ->
            ( { model | loading = True, error = Nothing }
            , API.getTestRuns model.procedureId model.limit model.offset RunsResponse
            )

        CreateResponse (Err error) ->
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
            [ Html.Attributes.class "page-header" ]
            [ Html.h1 [ Html.Attributes.class "mdc-typography--headline3" ] [ Html.text "Test Runs" ]
            , Html.button
                [ Html.Events.onClick SubmitCreate
                , Html.Attributes.class "mdc-button mdc-button--raised"
                , Html.Attributes.disabled model.loading
                ]
                [ Html.text "Create Run" ]
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
        , if model.loading && List.isEmpty model.runs then
            Html.div [] [ Html.text "Loading..." ]

          else
            viewRunsTable model.runs
        , viewPagination model
        ]


viewRunsTable : List TestRun -> Html Msg
viewRunsTable runs =
    Html.table
        [ Html.Attributes.class "mdc-data-table__table"
        , Html.Attributes.style "width" "100%"
        , Html.Attributes.style "border-collapse" "collapse"
        ]
        [ Html.thead []
            [ Html.tr []
                [ Html.th [ Html.Attributes.style "text-align" "left", Html.Attributes.style "padding" "12px" ] [ Html.text "Status" ]
                , Html.th [ Html.Attributes.style "text-align" "left", Html.Attributes.style "padding" "12px" ] [ Html.text "Version" ]
                , Html.th [ Html.Attributes.style "text-align" "left", Html.Attributes.style "padding" "12px" ] [ Html.text "Created" ]
                , Html.th [ Html.Attributes.style "text-align" "left", Html.Attributes.style "padding" "12px" ] [ Html.text "Actions" ]
                ]
            ]
        , Html.tbody []
            (List.map viewRunRow runs)
        ]


viewRunRow : TestRun -> Html Msg
viewRunRow run =
    Html.tr [ Html.Attributes.style "border-bottom" "1px solid #ddd" ]
        [ Html.td [ Html.Attributes.style "padding" "12px" ] [ Html.text (statusToString run.status) ]
        , Html.td [ Html.Attributes.style "padding" "12px" ] [ Html.text ("v" ++ String.fromInt run.procedureVersion) ]
        , Html.td [ Html.Attributes.style "padding" "12px" ] [ Html.text (formatTime run.createdAt) ]
        , Html.td [ Html.Attributes.style "padding" "12px" ]
            [ Html.a
                [ Html.Attributes.href ("/runs/" ++ run.id)
                , Html.Attributes.class "mdc-button"
                ]
                [ Html.text "View" ]
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


statusToString : TestRunStatus -> String
statusToString status =
    case status of
        Types.Pending ->
            "Pending"

        Types.Running ->
            "Running"

        Types.Passed ->
            "Passed"

        Types.Failed ->
            "Failed"

        Types.Skipped ->
            "Skipped"
