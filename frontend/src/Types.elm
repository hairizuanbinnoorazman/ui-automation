module Types exposing (..)

import Iso8601
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
    { name : String
    , instructions : String
    , imagePaths : List String
    }


type alias TestProcedureInput =
    { name : String
    , description : String
    , steps : List TestStep
    }


type alias DraftDiff =
    { draft : Maybe TestProcedure
    , committed : Maybe TestProcedure
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
    , procedureVersion : Int
    }


type alias TestRunStepNote =
    { id : String
    , testRunId : String
    , stepIndex : Int
    , notes : String
    , createdAt : Time.Posix
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
    , stepIndex : Maybe Int
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
    Decode.map3 TestStep
        (Decode.field "name" Decode.string)
        (Decode.field "instructions" Decode.string)
        (Decode.field "image_paths"
            (Decode.oneOf [ Decode.list Decode.string, Decode.null [] ])
        )


testProcedureDecoder : Decoder TestProcedure
testProcedureDecoder =
    Decode.map8
        (\id projectId name description steps version isLatest parentId ->
            \createdAt updatedAt deletedAt ->
                TestProcedure id projectId name description steps version isLatest parentId createdAt updatedAt deletedAt
        )
        (Decode.field "id" Decode.string)
        (Decode.field "project_id" Decode.string)
        (Decode.field "name" Decode.string)
        (Decode.field "description" Decode.string)
        (Decode.field "steps" (Decode.list testStepDecoder))
        (Decode.field "version" Decode.int)
        (Decode.field "is_latest" Decode.bool)
        (Decode.maybe (Decode.field "parent_id" Decode.string))
        |> Decode.andThen
            (\fn ->
                Decode.map3 fn
                    (Decode.field "created_at" timeDecoder)
                    (Decode.field "updated_at" timeDecoder)
                    (Decode.maybe (Decode.field "deleted_at" timeDecoder))
            )


draftDiffDecoder : Decoder DraftDiff
draftDiffDecoder =
    Decode.map2 DraftDiff
        (Decode.maybe (Decode.field "draft" testProcedureDecoder))
        (Decode.maybe (Decode.field "committed" testProcedureDecoder))


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
    Decode.map8
        (\id testProcedureId status notes startedAt completedAt createdAt updatedAt ->
            \procedureVersion ->
                TestRun id testProcedureId status notes startedAt completedAt createdAt updatedAt procedureVersion
        )
        (Decode.field "id" Decode.string)
        (Decode.field "test_procedure_id" Decode.string)
        (Decode.field "status" testRunStatusDecoder)
        (Decode.field "notes" Decode.string)
        (Decode.maybe (Decode.field "started_at" timeDecoder))
        (Decode.maybe (Decode.field "completed_at" timeDecoder))
        (Decode.field "created_at" timeDecoder)
        (Decode.field "updated_at" timeDecoder)
        |> Decode.andThen
            (\fn ->
                Decode.map fn
                    (Decode.oneOf [ Decode.field "procedure_version" Decode.int, Decode.succeed 0 ])
            )


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


testRunStepNoteDecoder : Decoder TestRunStepNote
testRunStepNoteDecoder =
    Decode.map5 TestRunStepNote
        (Decode.field "id" Decode.string)
        (Decode.field "test_run_id" Decode.string)
        (Decode.field "step_index" Decode.int)
        (Decode.field "notes" Decode.string)
        (Decode.field "created_at" timeDecoder)


testRunAssetDecoder : Decoder TestRunAsset
testRunAssetDecoder =
    Decode.map8 TestRunAsset
        (Decode.field "id" Decode.string)
        (Decode.field "test_run_id" Decode.string)
        (Decode.field "asset_type" assetTypeDecoder)
        (Decode.field "file_name" Decode.string)
        (Decode.field "asset_path" Decode.string)
        (Decode.oneOf [ Decode.field "description" Decode.string, Decode.succeed "" ])
        (Decode.field "uploaded_at" timeDecoder)
        (Decode.maybe (Decode.field "step_index" Decode.int))


paginatedDecoder : Decoder a -> Decoder (PaginatedResponse a)
paginatedDecoder itemDecoder =
    Decode.map4 PaginatedResponse
        (Decode.field "items" (Decode.oneOf [ Decode.list itemDecoder, Decode.null [] ]))
        (Decode.field "total" Decode.int)
        (Decode.field "limit" Decode.int)
        (Decode.field "offset" Decode.int)


timeDecoder : Decoder Time.Posix
timeDecoder =
    Iso8601.decoder



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
        [ ( "name", Encode.string step.name )
        , ( "instructions", Encode.string step.instructions )
        , ( "image_paths", Encode.list Encode.string step.imagePaths )
        ]


testProcedureInputEncoder : TestProcedureInput -> Encode.Value
testProcedureInputEncoder input =
    Encode.object
        [ ( "name", Encode.string input.name )
        , ( "description", Encode.string input.description )
        , ( "steps", Encode.list testStepEncoder input.steps )
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
