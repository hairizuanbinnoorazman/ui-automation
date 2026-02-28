module API exposing (..)

import File exposing (File)
import Http
import Json.Decode as Decode
import Json.Encode as Encode
import Types exposing (..)


baseUrl : String
baseUrl =
    "/api/v1"



-- Authentication


register : RegisterCredentials -> (Result Http.Error User -> msg) -> Cmd msg
register creds toMsg =
    Http.post
        { url = baseUrl ++ "/auth/register"
        , body = Http.jsonBody (registerCredentialsEncoder creds)
        , expect = Http.expectJson toMsg userDecoder
        }


login : LoginCredentials -> (Result Http.Error User -> msg) -> Cmd msg
login creds toMsg =
    Http.post
        { url = baseUrl ++ "/auth/login"
        , body = Http.jsonBody (loginCredentialsEncoder creds)
        , expect = Http.expectJson toMsg userDecoder
        }


logout : (Result Http.Error () -> msg) -> Cmd msg
logout toMsg =
    Http.post
        { url = baseUrl ++ "/auth/logout"
        , body = Http.emptyBody
        , expect = Http.expectWhatever toMsg
        }


getMe : (Result Http.Error User -> msg) -> Cmd msg
getMe toMsg =
    Http.get
        { url = baseUrl ++ "/auth/me"
        , expect = Http.expectJson toMsg userDecoder
        }



-- Projects


getProjects : Int -> Int -> (Result Http.Error (PaginatedResponse Project) -> msg) -> Cmd msg
getProjects limit offset toMsg =
    Http.get
        { url = baseUrl ++ "/projects?limit=" ++ String.fromInt limit ++ "&offset=" ++ String.fromInt offset
        , expect = Http.expectJson toMsg (paginatedDecoder projectDecoder)
        }


getProject : String -> (Result Http.Error Project -> msg) -> Cmd msg
getProject id toMsg =
    Http.get
        { url = baseUrl ++ "/projects/" ++ id
        , expect = Http.expectJson toMsg projectDecoder
        }


createProject : ProjectInput -> (Result Http.Error Project -> msg) -> Cmd msg
createProject input toMsg =
    Http.post
        { url = baseUrl ++ "/projects"
        , body = Http.jsonBody (projectInputEncoder input)
        , expect = Http.expectJson toMsg projectDecoder
        }


updateProject : String -> ProjectInput -> (Result Http.Error Project -> msg) -> Cmd msg
updateProject id input toMsg =
    Http.request
        { method = "PUT"
        , headers = []
        , url = baseUrl ++ "/projects/" ++ id
        , body = Http.jsonBody (projectInputEncoder input)
        , expect = Http.expectJson toMsg projectDecoder
        , timeout = Nothing
        , tracker = Nothing
        }


deleteProject : String -> (Result Http.Error () -> msg) -> Cmd msg
deleteProject id toMsg =
    Http.request
        { method = "DELETE"
        , headers = []
        , url = baseUrl ++ "/projects/" ++ id
        , body = Http.emptyBody
        , expect = Http.expectWhatever toMsg
        , timeout = Nothing
        , tracker = Nothing
        }



-- Test Procedures


getTestProcedures : String -> Int -> Int -> (Result Http.Error (PaginatedResponse TestProcedure) -> msg) -> Cmd msg
getTestProcedures projectId limit offset toMsg =
    Http.get
        { url = baseUrl ++ "/projects/" ++ projectId ++ "/procedures?limit=" ++ String.fromInt limit ++ "&offset=" ++ String.fromInt offset
        , expect = Http.expectJson toMsg (paginatedDecoder testProcedureDecoder)
        }


getTestProcedure : String -> String -> Bool -> (Result Http.Error TestProcedure -> msg) -> Cmd msg
getTestProcedure projectId procedureId isDraft toMsg =
    let
        draftParam =
            if isDraft then
                "?draft=true"

            else
                ""
    in
    Http.get
        { url = baseUrl ++ "/projects/" ++ projectId ++ "/procedures/" ++ procedureId ++ draftParam
        , expect = Http.expectJson toMsg testProcedureDecoder
        }


createTestProcedure : String -> TestProcedureInput -> (Result Http.Error TestProcedure -> msg) -> Cmd msg
createTestProcedure projectId input toMsg =
    Http.post
        { url = baseUrl ++ "/projects/" ++ projectId ++ "/procedures"
        , body = Http.jsonBody (testProcedureInputEncoder input)
        , expect = Http.expectJson toMsg testProcedureDecoder
        }


updateTestProcedure : String -> String -> TestProcedureInput -> (Result Http.Error TestProcedure -> msg) -> Cmd msg
updateTestProcedure projectId procedureId input toMsg =
    Http.request
        { method = "PUT"
        , headers = []
        , url = baseUrl ++ "/projects/" ++ projectId ++ "/procedures/" ++ procedureId
        , body = Http.jsonBody (testProcedureInputEncoder input)
        , expect = Http.expectJson toMsg testProcedureDecoder
        , timeout = Nothing
        , tracker = Nothing
        }


deleteTestProcedure : String -> String -> (Result Http.Error () -> msg) -> Cmd msg
deleteTestProcedure projectId procedureId toMsg =
    Http.request
        { method = "DELETE"
        , headers = []
        , url = baseUrl ++ "/projects/" ++ projectId ++ "/procedures/" ++ procedureId
        , body = Http.emptyBody
        , expect = Http.expectWhatever toMsg
        , timeout = Nothing
        , tracker = Nothing
        }


uploadStepImage : String -> File -> (Result Http.Error String -> msg) -> Cmd msg
uploadStepImage procedureId file toMsg =
    Http.post
        { url = baseUrl ++ "/procedures/" ++ procedureId ++ "/steps/images"
        , body = Http.multipartBody [ Http.filePart "image" file ]
        , expect = Http.expectJson toMsg (Decode.field "image_path" Decode.string)
        }


getDraftDiff : String -> (Result Http.Error DraftDiff -> msg) -> Cmd msg
getDraftDiff procedureId toMsg =
    Http.get
        { url = baseUrl ++ "/procedures/" ++ procedureId ++ "/diff"
        , expect = Http.expectJson toMsg draftDiffDecoder
        }


resetDraft : String -> (Result Http.Error () -> msg) -> Cmd msg
resetDraft procedureId toMsg =
    Http.post
        { url = baseUrl ++ "/procedures/" ++ procedureId ++ "/draft/reset"
        , body = Http.emptyBody
        , expect = Http.expectWhatever toMsg
        }


commitDraft : String -> (Result Http.Error TestProcedure -> msg) -> Cmd msg
commitDraft procedureId toMsg =
    Http.post
        { url = baseUrl ++ "/procedures/" ++ procedureId ++ "/draft/commit"
        , body = Http.emptyBody
        , expect = Http.expectJson toMsg testProcedureDecoder
        }


createProcedureVersion : String -> String -> (Result Http.Error TestProcedure -> msg) -> Cmd msg
createProcedureVersion projectId procedureId toMsg =
    Http.post
        { url = baseUrl ++ "/projects/" ++ projectId ++ "/procedures/" ++ procedureId ++ "/versions"
        , body = Http.emptyBody
        , expect = Http.expectJson toMsg testProcedureDecoder
        }


getProcedureVersions : String -> String -> (Result Http.Error (List TestProcedure) -> msg) -> Cmd msg
getProcedureVersions projectId procedureId toMsg =
    Http.get
        { url = baseUrl ++ "/projects/" ++ projectId ++ "/procedures/" ++ procedureId ++ "/versions"
        , expect = Http.expectJson toMsg (Decode.list testProcedureDecoder)
        }



-- Users


searchUsers : String -> (Result Http.Error UserListResponse -> msg) -> Cmd msg
searchUsers query toMsg =
    Http.get
        { url = baseUrl ++ "/users?search=" ++ query ++ "&limit=10"
        , expect = Http.expectJson toMsg userListResponseDecoder
        }


getUserById : String -> (Result Http.Error User -> msg) -> Cmd msg
getUserById userId toMsg =
    Http.get
        { url = baseUrl ++ "/users/" ++ userId
        , expect = Http.expectJson toMsg userDecoder
        }



-- Test Runs


getTestRuns : String -> Int -> Int -> (Result Http.Error (PaginatedResponse TestRun) -> msg) -> Cmd msg
getTestRuns procedureId limit offset toMsg =
    Http.get
        { url = baseUrl ++ "/procedures/" ++ procedureId ++ "/runs?limit=" ++ String.fromInt limit ++ "&offset=" ++ String.fromInt offset
        , expect = Http.expectJson toMsg (paginatedDecoder testRunDecoder)
        }


getTestRun : String -> (Result Http.Error TestRun -> msg) -> Cmd msg
getTestRun runId toMsg =
    Http.get
        { url = baseUrl ++ "/runs/" ++ runId
        , expect = Http.expectJson toMsg testRunDecoder
        }


createTestRun : String -> (Result Http.Error TestRun -> msg) -> Cmd msg
createTestRun procedureId toMsg =
    Http.post
        { url = baseUrl ++ "/procedures/" ++ procedureId ++ "/runs"
        , body = Http.emptyBody
        , expect = Http.expectJson toMsg testRunDecoder
        }


updateTestRun : String -> String -> (Result Http.Error TestRun -> msg) -> Cmd msg
updateTestRun runId notes toMsg =
    Http.request
        { method = "PUT"
        , headers = []
        , url = baseUrl ++ "/runs/" ++ runId
        , body = Http.jsonBody (Encode.object [ ( "notes", Encode.string notes ) ])
        , expect = Http.expectJson toMsg testRunDecoder
        , timeout = Nothing
        , tracker = Nothing
        }


assignTestRunUser : String -> String -> (Result Http.Error TestRun -> msg) -> Cmd msg
assignTestRunUser runId userId toMsg =
    Http.request
        { method = "PUT"
        , headers = []
        , url = baseUrl ++ "/runs/" ++ runId
        , body = Http.jsonBody (Encode.object [ ( "assigned_to", Encode.string userId ) ])
        , expect = Http.expectJson toMsg testRunDecoder
        , timeout = Nothing
        , tracker = Nothing
        }


unassignTestRunUser : String -> (Result Http.Error TestRun -> msg) -> Cmd msg
unassignTestRunUser runId toMsg =
    Http.request
        { method = "PUT"
        , headers = []
        , url = baseUrl ++ "/runs/" ++ runId
        , body = Http.jsonBody (Encode.object [ ( "assigned_to", Encode.string "" ) ])
        , expect = Http.expectJson toMsg testRunDecoder
        , timeout = Nothing
        , tracker = Nothing
        }


startTestRun : String -> (Result Http.Error TestRun -> msg) -> Cmd msg
startTestRun runId toMsg =
    Http.post
        { url = baseUrl ++ "/runs/" ++ runId ++ "/start"
        , body = Http.emptyBody
        , expect = Http.expectJson toMsg testRunDecoder
        }


completeTestRun : String -> CompleteTestRunInput -> (Result Http.Error TestRun -> msg) -> Cmd msg
completeTestRun runId input toMsg =
    Http.post
        { url = baseUrl ++ "/runs/" ++ runId ++ "/complete"
        , body = Http.jsonBody (completeTestRunInputEncoder input)
        , expect = Http.expectJson toMsg testRunDecoder
        }



-- Test Run Assets


getTestRunAssets : String -> (Result Http.Error (List TestRunAsset) -> msg) -> Cmd msg
getTestRunAssets runId toMsg =
    Http.get
        { url = baseUrl ++ "/runs/" ++ runId ++ "/assets"
        , expect = Http.expectJson toMsg (Decode.list testRunAssetDecoder)
        }


deleteTestRunAsset : String -> String -> (Result Http.Error () -> msg) -> Cmd msg
deleteTestRunAsset runId assetId toMsg =
    Http.request
        { method = "DELETE"
        , headers = []
        , url = baseUrl ++ "/runs/" ++ runId ++ "/assets/" ++ assetId
        , body = Http.emptyBody
        , expect = Http.expectWhatever toMsg
        , timeout = Nothing
        , tracker = Nothing
        }


getAssetDownloadUrl : String -> String -> String
getAssetDownloadUrl runId assetId =
    baseUrl ++ "/runs/" ++ runId ++ "/assets/" ++ assetId


getRunProcedure : String -> (Result Http.Error TestProcedure -> msg) -> Cmd msg
getRunProcedure runId toMsg =
    Http.get
        { url = baseUrl ++ "/runs/" ++ runId ++ "/procedure"
        , expect = Http.expectJson toMsg testProcedureDecoder
        }


getStepNotes : String -> (Result Http.Error (List TestRunStepNote) -> msg) -> Cmd msg
getStepNotes runId toMsg =
    Http.get
        { url = baseUrl ++ "/runs/" ++ runId ++ "/steps/notes"
        , expect = Http.expectJson toMsg (Decode.list testRunStepNoteDecoder)
        }


setStepNote : String -> Int -> String -> (Result Http.Error TestRunStepNote -> msg) -> Cmd msg
setStepNote runId stepIndex notes toMsg =
    Http.request
        { method = "PUT"
        , headers = []
        , url = baseUrl ++ "/runs/" ++ runId ++ "/steps/" ++ String.fromInt stepIndex ++ "/notes"
        , body = Http.jsonBody (Encode.object [ ( "notes", Encode.string notes ) ])
        , expect = Http.expectJson toMsg testRunStepNoteDecoder
        , timeout = Nothing
        , tracker = Nothing
        }


uploadStepAsset : String -> Int -> File -> (Result Http.Error TestRunAsset -> msg) -> Cmd msg
uploadStepAsset runId stepIndex file toMsg =
    Http.post
        { url = baseUrl ++ "/runs/" ++ runId ++ "/assets"
        , body =
            Http.multipartBody
                [ Http.filePart "file" file
                , Http.stringPart "asset_type" "image"
                , Http.stringPart "step_index" (String.fromInt stepIndex)
                ]
        , expect = Http.expectJson toMsg testRunAssetDecoder
        }



-- Endpoints


getEndpoints : Int -> Int -> (Result Http.Error (PaginatedResponse Endpoint) -> msg) -> Cmd msg
getEndpoints limit offset toMsg =
    Http.get
        { url = baseUrl ++ "/endpoints?limit=" ++ String.fromInt limit ++ "&offset=" ++ String.fromInt offset
        , expect = Http.expectJson toMsg (paginatedDecoder endpointDecoder)
        }


getEndpoint : String -> (Result Http.Error Endpoint -> msg) -> Cmd msg
getEndpoint id toMsg =
    Http.get
        { url = baseUrl ++ "/endpoints/" ++ id
        , expect = Http.expectJson toMsg endpointDecoder
        }


createEndpoint : EndpointInput -> (Result Http.Error Endpoint -> msg) -> Cmd msg
createEndpoint input toMsg =
    Http.post
        { url = baseUrl ++ "/endpoints"
        , body = Http.jsonBody (endpointInputEncoder input)
        , expect = Http.expectJson toMsg endpointDecoder
        }


updateEndpoint : String -> EndpointInput -> (Result Http.Error Endpoint -> msg) -> Cmd msg
updateEndpoint id input toMsg =
    Http.request
        { method = "PUT"
        , headers = []
        , url = baseUrl ++ "/endpoints/" ++ id
        , body = Http.jsonBody (endpointInputEncoder input)
        , expect = Http.expectJson toMsg endpointDecoder
        , timeout = Nothing
        , tracker = Nothing
        }


deleteEndpoint : String -> (Result Http.Error () -> msg) -> Cmd msg
deleteEndpoint id toMsg =
    Http.request
        { method = "DELETE"
        , headers = []
        , url = baseUrl ++ "/endpoints/" ++ id
        , body = Http.emptyBody
        , expect = Http.expectWhatever toMsg
        , timeout = Nothing
        , tracker = Nothing
        }



-- Jobs


getJobs : Int -> Int -> (Result Http.Error (PaginatedResponse Job) -> msg) -> Cmd msg
getJobs limit offset toMsg =
    Http.get
        { url = baseUrl ++ "/jobs?limit=" ++ String.fromInt limit ++ "&offset=" ++ String.fromInt offset
        , expect = Http.expectJson toMsg (paginatedDecoder jobDecoder)
        }


getJob : String -> (Result Http.Error Job -> msg) -> Cmd msg
getJob id toMsg =
    Http.get
        { url = baseUrl ++ "/jobs/" ++ id
        , expect = Http.expectJson toMsg jobDecoder
        }


createJob : CreateJobInput -> (Result Http.Error Job -> msg) -> Cmd msg
createJob input toMsg =
    Http.post
        { url = baseUrl ++ "/jobs"
        , body = Http.jsonBody (createJobInputEncoder input)
        , expect = Http.expectJson toMsg jobDecoder
        }


stopJob : String -> (Result Http.Error Job -> msg) -> Cmd msg
stopJob id toMsg =
    Http.post
        { url = baseUrl ++ "/jobs/" ++ id ++ "/stop"
        , body = Http.emptyBody
        , expect = Http.expectJson toMsg jobDecoder
        }



-- API Tokens


getAPITokens : (Result Http.Error TokenListResponse -> msg) -> Cmd msg
getAPITokens toMsg =
    Http.get
        { url = baseUrl ++ "/tokens"
        , expect = Http.expectJson toMsg tokenListResponseDecoder
        }


createAPIToken : CreateTokenInput -> (Result Http.Error CreateTokenResponse -> msg) -> Cmd msg
createAPIToken input toMsg =
    Http.post
        { url = baseUrl ++ "/tokens"
        , body = Http.jsonBody (createTokenInputEncoder input)
        , expect = Http.expectJson toMsg createTokenResponseDecoder
        }


revokeAPIToken : String -> (Result Http.Error () -> msg) -> Cmd msg
revokeAPIToken tokenId toMsg =
    Http.request
        { method = "DELETE"
        , headers = []
        , url = baseUrl ++ "/tokens/" ++ tokenId
        , body = Http.emptyBody
        , expect = Http.expectWhatever toMsg
        , timeout = Nothing
        , tracker = Nothing
        }
