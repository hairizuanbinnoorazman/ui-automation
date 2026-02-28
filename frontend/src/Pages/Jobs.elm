module Pages.Jobs exposing (Model, Msg, init, update, view)

import API
import Html exposing (Html)
import Html.Attributes
import Html.Events
import Http
import Json.Encode as Encode
import Time
import Types exposing (Job, JobStatus(..), PaginatedResponse)



-- MODEL


type alias Model =
    { jobs : List Job
    , total : Int
    , limit : Int
    , offset : Int
    , loading : Bool
    , error : Maybe String
    , selectedJob : Maybe Job
    }


init : ( Model, Cmd Msg )
init =
    ( { jobs = []
      , total = 0
      , limit = 10
      , offset = 0
      , loading = True
      , error = Nothing
      , selectedJob = Nothing
      }
    , API.getJobs 10 0 JobsResponse
    )



-- UPDATE


type Msg
    = JobsResponse (Result Http.Error (PaginatedResponse Job))
    | LoadPage Int
    | SelectJob Job
    | ClearSelection
    | StopJob String
    | StopJobResponse (Result Http.Error Job)


update : Msg -> Model -> ( Model, Cmd Msg )
update msg model =
    case msg of
        JobsResponse (Ok response) ->
            ( { model
                | jobs = response.items
                , total = response.total
                , loading = False
                , error = Nothing
              }
            , Cmd.none
            )

        JobsResponse (Err error) ->
            ( { model
                | loading = False
                , error = Just (httpErrorToString error)
              }
            , Cmd.none
            )

        LoadPage offset ->
            ( { model | loading = True, offset = offset, selectedJob = Nothing }
            , API.getJobs model.limit offset JobsResponse
            )

        SelectJob job ->
            ( { model | selectedJob = Just job }
            , Cmd.none
            )

        ClearSelection ->
            ( { model | selectedJob = Nothing }
            , Cmd.none
            )

        StopJob id ->
            ( { model | loading = True }
            , API.stopJob id StopJobResponse
            )

        StopJobResponse (Ok job) ->
            ( { model
                | loading = False
                , selectedJob = Just job
              }
            , API.getJobs model.limit model.offset JobsResponse
            )

        StopJobResponse (Err error) ->
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
            [ Html.h1 [ Html.Attributes.class "mdc-typography--headline3" ] [ Html.text "Jobs" ]
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
        , if model.loading && List.isEmpty model.jobs then
            Html.div [] [ Html.text "Loading..." ]

          else
            viewJobsTable model.jobs model.selectedJob
        , case model.selectedJob of
            Just job ->
                viewJobDetail job

            Nothing ->
                Html.text ""
        , viewPagination model
        ]


viewJobsTable : List Job -> Maybe Job -> Html Msg
viewJobsTable jobs selectedJob =
    Html.table
        [ Html.Attributes.class "mdc-data-table__table"
        , Html.Attributes.style "width" "100%"
        , Html.Attributes.style "border-collapse" "collapse"
        ]
        [ Html.thead []
            [ Html.tr []
                [ Html.th [ Html.Attributes.style "text-align" "left", Html.Attributes.style "padding" "12px" ] [ Html.text "Type" ]
                , Html.th [ Html.Attributes.style "text-align" "left", Html.Attributes.style "padding" "12px" ] [ Html.text "Status" ]
                , Html.th [ Html.Attributes.style "text-align" "left", Html.Attributes.style "padding" "12px" ] [ Html.text "Start Time" ]
                , Html.th [ Html.Attributes.style "text-align" "left", Html.Attributes.style "padding" "12px" ] [ Html.text "Duration" ]
                , Html.th [ Html.Attributes.style "text-align" "left", Html.Attributes.style "padding" "12px" ] [ Html.text "Created" ]
                ]
            ]
        , Html.tbody []
            (List.map (viewJobRow selectedJob) jobs)
        ]


viewJobRow : Maybe Job -> Job -> Html Msg
viewJobRow selectedJob job =
    let
        isSelected =
            case selectedJob of
                Just selected ->
                    selected.id == job.id

                Nothing ->
                    False

        rowBg =
            if isSelected then
                "#e8eaf6"

            else
                "transparent"
    in
    Html.tr
        [ Html.Attributes.style "border-bottom" "1px solid #ddd"
        , Html.Attributes.style "cursor" "pointer"
        , Html.Attributes.style "background-color" rowBg
        , Html.Events.onClick (SelectJob job)
        ]
        [ Html.td [ Html.Attributes.style "padding" "12px" ] [ Html.text job.jobType ]
        , Html.td [ Html.Attributes.style "padding" "12px" ] [ viewStatusBadge job.status ]
        , Html.td [ Html.Attributes.style "padding" "12px" ]
            [ Html.text
                (case job.startTime of
                    Just t ->
                        formatTime t

                    Nothing ->
                        "-"
                )
            ]
        , Html.td [ Html.Attributes.style "padding" "12px" ]
            [ Html.text
                (case job.duration of
                    Just d ->
                        formatDuration d

                    Nothing ->
                        "-"
                )
            ]
        , Html.td [ Html.Attributes.style "padding" "12px" ] [ Html.text (formatTime job.createdAt) ]
        ]


viewStatusBadge : JobStatus -> Html msg
viewStatusBadge status =
    let
        ( bgColor, textColor, label ) =
            case status of
                JobCreated ->
                    ( "#e0e0e0", "#616161", "created" )

                JobRunning ->
                    ( "#bbdefb", "#1565c0", "running" )

                JobStopped ->
                    ( "#ffe0b2", "#e65100", "stopped" )

                JobFailed ->
                    ( "#ffcdd2", "#c62828", "failed" )

                JobSuccess ->
                    ( "#c8e6c9", "#2e7d32", "success" )
    in
    Html.span
        [ Html.Attributes.style "display" "inline-block"
        , Html.Attributes.style "padding" "4px 12px"
        , Html.Attributes.style "border-radius" "12px"
        , Html.Attributes.style "font-size" "12px"
        , Html.Attributes.style "font-weight" "500"
        , Html.Attributes.style "background-color" bgColor
        , Html.Attributes.style "color" textColor
        ]
        [ Html.text label ]


viewJobDetail : Job -> Html Msg
viewJobDetail job =
    Html.div
        [ Html.Attributes.style "margin-top" "24px"
        , Html.Attributes.style "padding" "20px"
        , Html.Attributes.style "border" "1px solid #ddd"
        , Html.Attributes.style "border-radius" "4px"
        , Html.Attributes.style "background" "#fafafa"
        ]
        [ Html.div
            [ Html.Attributes.style "display" "flex"
            , Html.Attributes.style "justify-content" "space-between"
            , Html.Attributes.style "align-items" "center"
            , Html.Attributes.style "margin-bottom" "16px"
            ]
            [ Html.h3
                [ Html.Attributes.style "margin" "0"
                , Html.Attributes.class "mdc-typography--headline6"
                ]
                [ Html.text ("Job Detail: " ++ job.jobType) ]
            , Html.div
                [ Html.Attributes.style "display" "flex"
                , Html.Attributes.style "gap" "8px"
                ]
                [ case job.status of
                    JobRunning ->
                        Html.button
                            [ Html.Events.onClick (StopJob job.id)
                            , Html.Attributes.class "mdc-button mdc-button--raised"
                            , Html.Attributes.style "background-color" "#f44336"
                            ]
                            [ Html.text "Stop" ]

                    _ ->
                        Html.text ""
                , Html.button
                    [ Html.Events.onClick ClearSelection
                    , Html.Attributes.class "mdc-button"
                    ]
                    [ Html.text "Close" ]
                ]
            ]
        , Html.div [ Html.Attributes.style "margin-bottom" "16px" ]
            [ Html.strong [] [ Html.text "Status: " ]
            , viewStatusBadge job.status
            ]
        , Html.div [ Html.Attributes.style "margin-bottom" "16px" ]
            [ Html.strong [] [ Html.text "Config:" ]
            , Html.pre
                [ Html.Attributes.style "background" "#f5f5f5"
                , Html.Attributes.style "padding" "12px"
                , Html.Attributes.style "border-radius" "4px"
                , Html.Attributes.style "overflow-x" "auto"
                , Html.Attributes.style "font-size" "13px"
                , Html.Attributes.style "margin-top" "8px"
                ]
                [ Html.text (Encode.encode 2 job.config) ]
            ]
        , Html.div [ Html.Attributes.style "margin-bottom" "16px" ]
            [ Html.strong [] [ Html.text "Result:" ]
            , Html.pre
                [ Html.Attributes.style "background" "#f5f5f5"
                , Html.Attributes.style "padding" "12px"
                , Html.Attributes.style "border-radius" "4px"
                , Html.Attributes.style "overflow-x" "auto"
                , Html.Attributes.style "font-size" "13px"
                , Html.Attributes.style "margin-top" "8px"
                ]
                [ Html.text (Encode.encode 2 job.result) ]
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


formatDuration : Int -> String
formatDuration seconds =
    if seconds >= 3600 then
        String.fromInt (seconds // 3600) ++ "h " ++ String.fromInt (modBy 60 (seconds // 60)) ++ "m"

    else if seconds >= 60 then
        String.fromInt (seconds // 60) ++ "m " ++ String.fromInt (modBy 60 seconds) ++ "s"

    else
        String.fromInt seconds ++ "s"


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
