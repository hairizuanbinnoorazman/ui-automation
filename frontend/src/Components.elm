module Components exposing (viewDialogOverlay, viewFormField, viewSelectField, viewTextArea)

import Html exposing (Html)
import Html.Attributes


viewDialogOverlay : String -> List (Html msg) -> List (Html msg) -> Html msg
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
            , Html.Attributes.style "max-width" "600px"
            ]
            [ Html.h2
                [ Html.Attributes.class "mdc-typography--headline6"
                , Html.Attributes.style "margin-top" "0"
                , Html.Attributes.style "margin-bottom" "24px"
                ]
                [ Html.text title ]
            , Html.div
                [ Html.Attributes.style "margin-bottom" "24px"
                ]
                content
            , Html.div
                [ Html.Attributes.style "display" "flex"
                , Html.Attributes.style "justify-content" "flex-end"
                , Html.Attributes.style "gap" "8px"
                ]
                actions
            ]
        ]


viewFormField : String -> List (Html.Attribute msg) -> Html msg
viewFormField labelText inputAttrs =
    Html.div
        [ Html.Attributes.style "margin-bottom" "20px"
        ]
        [ Html.label
            [ Html.Attributes.style "display" "block"
            , Html.Attributes.style "margin-bottom" "8px"
            , Html.Attributes.style "font-weight" "500"
            , Html.Attributes.style "color" "#333"
            ]
            [ Html.text labelText ]
        , Html.input
            ([ Html.Attributes.style "width" "100%"
             , Html.Attributes.style "padding" "10px"
             , Html.Attributes.style "border" "1px solid #ddd"
             , Html.Attributes.style "border-radius" "4px"
             , Html.Attributes.style "font-size" "14px"
             , Html.Attributes.style "box-sizing" "border-box"
             ]
                ++ inputAttrs
            )
            []
        ]


viewTextArea : String -> List (Html.Attribute msg) -> Html msg
viewTextArea labelText textareaAttrs =
    Html.div
        [ Html.Attributes.style "margin-bottom" "20px"
        ]
        [ Html.label
            [ Html.Attributes.style "display" "block"
            , Html.Attributes.style "margin-bottom" "8px"
            , Html.Attributes.style "font-weight" "500"
            , Html.Attributes.style "color" "#333"
            ]
            [ Html.text labelText ]
        , Html.textarea
            ([ Html.Attributes.style "width" "100%"
             , Html.Attributes.style "padding" "10px"
             , Html.Attributes.style "border" "1px solid #ddd"
             , Html.Attributes.style "border-radius" "4px"
             , Html.Attributes.style "font-size" "14px"
             , Html.Attributes.style "box-sizing" "border-box"
             , Html.Attributes.style "min-height" "80px"
             , Html.Attributes.style "resize" "vertical"
             ]
                ++ textareaAttrs
            )
            []
        ]


viewSelectField : String -> List (Html.Attribute msg) -> List (Html msg) -> Html msg
viewSelectField labelText selectAttrs options =
    Html.div
        [ Html.Attributes.style "margin-bottom" "20px"
        ]
        [ Html.label
            [ Html.Attributes.style "display" "block"
            , Html.Attributes.style "margin-bottom" "8px"
            , Html.Attributes.style "font-weight" "500"
            , Html.Attributes.style "color" "#333"
            ]
            [ Html.text labelText ]
        , Html.select
            ([ Html.Attributes.style "width" "100%"
             , Html.Attributes.style "padding" "10px"
             , Html.Attributes.style "border" "1px solid #ddd"
             , Html.Attributes.style "border-radius" "4px"
             , Html.Attributes.style "font-size" "14px"
             , Html.Attributes.style "box-sizing" "border-box"
             ]
                ++ selectAttrs
            )
            options
        ]
