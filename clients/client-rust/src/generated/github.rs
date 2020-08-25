#![allow(unused_imports)]
#![cfg_attr(rustfmt, rustfmt_skip)]
/** THIS FILE IS AUTOMATICALLY GENERATED. DO NOT EDIT */
use crate::{Client, Credentials};
use failure::Error;
use serde_json::Value;
use crate::util::urlencode;

/// GitHub Service
///
/// The github service is responsible for creating tasks in response
/// to GitHub events, and posting results to the GitHub UI.
///
/// This document describes the API end-point for consuming GitHub
/// web hooks, as well as some useful consumer APIs.
///
/// When Github forbids an action, this service returns an HTTP 403
/// with code ForbiddenByGithub.
pub struct Github (Client);

#[allow(non_snake_case)]
impl Github {
    pub fn new(root_url: &str, credentials: Option<Credentials>) -> Result<Self, Error> {
        Ok(Self(Client::new(root_url, "github", "v1", credentials)?))
    }

    /// Ping Server
    /// 
    /// Respond without doing anything.
    /// This endpoint is used to check that the service is up.
    pub async fn ping(&self) -> Result<(), Error> {
        let method = "GET";
        let path = "ping";
        let query = None;
        let body = None;

        let resp = self.0.request(method, path, query, body).await?;

        resp.bytes().await?;
        Ok(())
    }

    /// Consume GitHub WebHook
    /// 
    /// Capture a GitHub event and publish it via pulse, if it's a push,
    /// release or pull request.
    pub async fn githubWebHookConsumer(&self) -> Result<(), Error> {
        let method = "POST";
        let path = "github";
        let query = None;
        let body = None;

        let resp = self.0.request(method, path, query, body).await?;

        resp.bytes().await?;
        Ok(())
    }

    /// List of Builds
    /// 
    /// A paginated list of builds that have been run in
    /// Taskcluster. Can be filtered on various git-specific
    /// fields.
    pub async fn builds(&self, continuationToken: Option<&str>, limit: Option<&str>, organization: Option<&str>, repository: Option<&str>, sha: Option<&str>) -> Result<Value, Error> {
        let method = "GET";
        let path = "builds";
        let mut query = None;
        if let Some(q) = continuationToken {
            query.get_or_insert_with(Vec::new).push(("continuationToken", q));
        }
        if let Some(q) = limit {
            query.get_or_insert_with(Vec::new).push(("limit", q));
        }
        if let Some(q) = organization {
            query.get_or_insert_with(Vec::new).push(("organization", q));
        }
        if let Some(q) = repository {
            query.get_or_insert_with(Vec::new).push(("repository", q));
        }
        if let Some(q) = sha {
            query.get_or_insert_with(Vec::new).push(("sha", q));
        }
        let body = None;

        let resp = self.0.request(method, path, query, body).await?;

        Ok(resp.json().await?)
    }

    /// Latest Build Status Badge
    /// 
    /// Checks the status of the latest build of a given branch
    /// and returns corresponding badge svg.
    pub async fn badge(&self, owner: &str, repo: &str, branch: &str) -> Result<(), Error> {
        let method = "GET";
        let path = format!("repository/{}/{}/{}/badge.svg", urlencode(owner), urlencode(repo), urlencode(branch));
        let query = None;
        let body = None;

        let resp = self.0.request(method, &path, query, body).await?;

        resp.bytes().await?;
        Ok(())
    }

    /// Get Repository Info
    /// 
    /// Returns any repository metadata that is
    /// useful within Taskcluster related services.
    pub async fn repository(&self, owner: &str, repo: &str) -> Result<Value, Error> {
        let method = "GET";
        let path = format!("repository/{}/{}", urlencode(owner), urlencode(repo));
        let query = None;
        let body = None;

        let resp = self.0.request(method, &path, query, body).await?;

        Ok(resp.json().await?)
    }

    /// Latest Status for Branch
    /// 
    /// For a given branch of a repository, this will always point
    /// to a status page for the most recent task triggered by that
    /// branch.
    /// 
    /// Note: This is a redirect rather than a direct link.
    pub async fn latest(&self, owner: &str, repo: &str, branch: &str) -> Result<(), Error> {
        let method = "GET";
        let path = format!("repository/{}/{}/{}/latest", urlencode(owner), urlencode(repo), urlencode(branch));
        let query = None;
        let body = None;

        let resp = self.0.request(method, &path, query, body).await?;

        resp.bytes().await?;
        Ok(())
    }

    /// Post a status against a given changeset
    /// 
    /// For a given changeset (SHA) of a repository, this will attach a "commit status"
    /// on github. These statuses are links displayed next to each revision.
    /// The status is either OK (green check) or FAILURE (red cross), 
    /// made of a custom title and link.
    pub async fn createStatus(&self, owner: &str, repo: &str, sha: &str, payload: &Value) -> Result<(), Error> {
        let method = "POST";
        let path = format!("repository/{}/{}/statuses/{}", urlencode(owner), urlencode(repo), urlencode(sha));
        let query = None;
        let body = Some(payload);

        let resp = self.0.request(method, &path, query, body).await?;

        resp.bytes().await?;
        Ok(())
    }

    /// Post a comment on a given GitHub Issue or Pull Request
    /// 
    /// For a given Issue or Pull Request of a repository, this will write a new message.
    pub async fn createComment(&self, owner: &str, repo: &str, number: &str, payload: &Value) -> Result<(), Error> {
        let method = "POST";
        let path = format!("repository/{}/{}/issues/{}/comments", urlencode(owner), urlencode(repo), urlencode(number));
        let query = None;
        let body = Some(payload);

        let resp = self.0.request(method, &path, query, body).await?;

        resp.bytes().await?;
        Ok(())
    }
}