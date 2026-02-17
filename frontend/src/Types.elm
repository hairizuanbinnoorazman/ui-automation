module Types exposing (..)

import Json.Decode as Decode exposing (Decoder)
import Json.Encode as Encode
import Time


-- User Types


type alias User =
    { id : String
    , email : String
    , username : String
    , createdAt : Time.Posix
    , updatedAt : Time.Posix
    }


type alias LoginCredentials =
    { email : String
    , password : String
    }


type alias RegisterCredentials =
    { email : String
    , username : String
    , password : String
    }



-- Project Types


type alias Project =
    { id : String
    , name : String
    , description : String
    , ownerId : String
    , createdAt : Time.Posix
    , updatedAt : Time.Posix
    , deletedAt : Maybe Time.Posix
    }


type alias ProjectInput =
    { name : String
    , description : String
    }



-- Test Procedure Types


type alias TestProcedure =
    { id : String
    , projectId : String
    , name : String
    , description : String
    , steps : List TestStep
    , version : Int
    , isLatest : Bool
    , parentId : Maybe String
    , createdAt : Time.Posix
    , updatedAt : Time.Posix
    , deletedAt : Maybe Time.Posix
    }


type alias TestStep =
    { action : String
    , selector : Maybe String
    , url : Maybe String
    , value : Maybe String
    , endpoint : Maybe String
    }


type alias TestProcedureInput =
    { name : String
    , description : String
    , steps : List TestStep
    }



-- Test Run Types


type TestRunStatus
    = Pending
    | Running
    | Passed
    | Failed
    | Skipped


type alias TestRun =
    { id : String
    , testProcedureId : String
    , status : TestRunStatus
    , notes : String
    , startedAt : Maybe Time.Posix
    , completedAt : Maybe Time.Posix
    , createdAt : Time.Posix
    , updatedAt : Time.Posix
    }


type alias TestRunInput =
    { notes : String
    }


type alias CompleteTestRunInput =
    { status : TestRunStatus
    , notes : String
    }



-- Test Run Asset Types


type AssetType
    = Image
    | Video
    | Binary
    | Document


type alias TestRunAsset =
    { id : String
    , testRunId : String
    , assetType : AssetType
    , filename : String
    , filepath : String
    , description : String
    , createdAt : Time.Posix
    }



-- Pagination


type alias PaginatedResponse a =
    { items : List a
    , total : Int
    , limit : Int
    , offset : Int
    }



-- JSON Decoders


userDecoder : Decoder User
userDecoder =
    Decode.map5 User
        (Decode.field "id" Decode.string)
        (Decode.field "email" Decode.string)
        (Decode.field "username" Decode.string)
        (Decode.field "created_at" timeDecoder)
        (Decode.field "updated_at" timeDecoder)


projectDecoder : Decoder Project
projectDecoder =
    Decode.map7 Project
        (Decode.field "id" Decode.string)
        (Decode.field "name" Decode.string)
        (Decode.field "description" Decode.string)
        (Decode.field "owner_id" Decode.string)
        (Decode.field "created_at" timeDecoder)
        (Decode.field "updated_at" timeDecoder)
        (Decode.maybe (Decode.field "deleted_at" timeDecoder))


testStepDecoder : Decoder TestStep
testStepDecoder =
    Decode.map5 TestStep
        (Decode.field "action" Decode.string)
        (Decode.maybe (Decode.field "selector" Decode.string))
        (Decode.maybe (Decode.field "url" Decode.string))
        (Decode.maybe (Decode.field "value" Decode.string))
        (Decode.maybe (Decode.field "endpoint" Decode.string))


testProcedureDecoder : Decoder TestProcedure
testProcedureDecoder =
    Decode.map10 TestProcedure
        (Decode.field "id" Decode.string)
        (Decode.field "project_id" Decode.string)
        (Decode.field "name" Decode.string)
        (Decode.field "description" Decode.string)
        (Decode.field "steps" (Decode.list testStepDecoder))
        (Decode.field "version" Decode.int)
        (Decode.field "is_latest" Decode.bool)
        (Decode.maybe (Decode.field "parent_id" Decode.string))
        (Decode.field "created_at" timeDecoder)
        (Decode.field "updated_at" timeDecoder)
        (Decode.maybe (Decode.field "deleted_at" timeDecoder))


testRunStatusDecoder : Decoder TestRunStatus
testRunStatusDecoder =
    Decode.string
        |> Decode.andThen
            (\str ->
                case str of
                    "pending" ->
                        Decode.succeed Pending

                    "running" ->
                        Decode.succeed Running

                    "passed" ->
                        Decode.succeed Passed

                    "failed" ->
                        Decode.succeed Failed

                    "skipped" ->
                        Decode.succeed Skipped

                    _ ->
                        Decode.fail ("Unknown status: " ++ str)
            )


testRunDecoder : Decoder TestRun
testRunDecoder =
    Decode.map8 TestRun
        (Decode.field "id" Decode.string)
        (Decode.field "test_procedure_id" Decode.string)
        (Decode.field "status" testRunStatusDecoder)
        (Decode.field "notes" Decode.string)
        (Decode.maybe (Decode.field "started_at" timeDecoder))
        (Decode.maybe (Decode.field "completed_at" timeDecoder))
        (Decode.field "created_at" timeDecoder)
        (Decode.field "updated_at" timeDecoder)


assetTypeDecoder : Decoder AssetType
assetTypeDecoder =
    Decode.string
        |> Decode.andThen
            (\str ->
                case str of
                    "image" ->
                        Decode.succeed Image

                    "video" ->
                        Decode.succeed Video

                    "binary" ->
                        Decode.succeed Binary

                    "document" ->
                        Decode.succeed Document

                    _ ->
                        Decode.fail ("Unknown asset type: " ++ str)
            )


testRunAssetDecoder : Decoder TestRunAsset
testRunAssetDecoder =
    Decode.map6 TestRunAsset
        (Decode.field "id" Decode.string)
        (Decode.field "test_run_id" Decode.string)
        (Decode.field "asset_type" assetTypeDecoder)
        (Decode.field "filename" Decode.string)
        (Decode.field "filepath" Decode.string)
        (Decode.field "description" Decode.string)
        (Decode.field "created_at" timeDecoder)


paginatedDecoder : Decoder a -> Decoder (PaginatedResponse a)
paginatedDecoder itemDecoder =
    Decode.map4 PaginatedResponse
        (Decode.field "items" (Decode.list itemDecoder))
        (Decode.field "total" Decode.int)
        (Decode.field "limit" Decode.int)
        (Decode.field "offset" Decode.int)


timeDecoder : Decoder Time.Posix
timeDecoder =
    Decode.string
        |> Decode.andThen
            (\str ->
                case String.toInt str of
                    Just ms ->
                        Decode.succeed (Time.millisToPosix ms)

                    Nothing ->
                        Decode.fail "Invalid timestamp"
            )



-- JSON Encoders


loginCredentialsEncoder : LoginCredentials -> Encode.Value
loginCredentialsEncoder creds =
    Encode.object
        [ ( "email", Encode.string creds.email )
        , ( "password", Encode.string creds.password )
        ]


registerCredentialsEncoder : RegisterCredentials -> Encode.Value
registerCredentialsEncoder creds =
    Encode.object
        [ ( "email", Encode.string creds.email )
        , ( "username", Encode.string creds.username )
        , ( "password", Encode.string creds.password )
        ]


projectInputEncoder : ProjectInput -> Encode.Value
projectInputEncoder input =
    Encode.object
        [ ( "name", Encode.string input.name )
        , ( "description", Encode.string input.description )
        ]


testStepEncoder : TestStep -> Encode.Value
testStepEncoder step =
    Encode.object
        (List.filterMap identity
            [ Just ( "action", Encode.string step.action )
            , Maybe.map (\s -> ( "selector", Encode.string s )) step.selector
            , Maybe.map (\u -> ( "url", Encode.string u )) step.url
            , Maybe.map (\v -> ( "value", Encode.string v )) step.value
            , Maybe.map (\e -> ( "endpoint", Encode.string e )) step.endpoint
            ]
        )


testProcedureInputEncoder : TestProcedureInput -> Encode.Value
testProcedureInputEncoder input =
    Encode.object
        [ ( "name", Encode.string input.name )
        , ( "description", Encode.string input.description )
        , ( "steps", Encode.list testStepEncoder input.steps )
        ]


testRunInputEncoder : TestRunInput -> Encode.Value
testRunInputEncoder input =
    Encode.object
        [ ( "notes", Encode.string input.notes )
        ]


testRunStatusToString : TestRunStatus -> String
testRunStatusToString status =
    case status of
        Pending ->
            "pending"

        Running ->
            "running"

        Passed ->
            "passed"

        Failed ->
            "failed"

        Skipped ->
            "skipped"


completeTestRunInputEncoder : CompleteTestRunInput -> Encode.Value
completeTestRunInputEncoder input =
    Encode.object
        [ ( "status", Encode.string (testRunStatusToString input.status) )
        , ( "notes", Encode.string input.notes )
        ]


assetTypeToString : AssetType -> String
assetTypeToString assetType =
    case assetType of
        Image ->
            "image"

        Video ->
            "video"

        Binary ->
            "binary"

        Document ->
            "document"
