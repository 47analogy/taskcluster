#![allow(unused_imports)]
#![cfg_attr(rustfmt, rustfmt_skip)]
/** THIS FILE IS AUTOMATICALLY GENERATED. DO NOT EDIT */
use crate::{Client, Credentials};
use failure::Error;
use serde_json::Value;
use crate::util::urlencode;

/// Auth Service
///
/// Authentication related API end-points for Taskcluster and related
/// services. These API end-points are of interest if you wish to:
///   * Authorize a request signed with Taskcluster credentials,
///   * Manage clients and roles,
///   * Inspect or audit clients and roles,
///   * Gain access to various services guarded by this API.
///
pub struct Auth (Client);

#[allow(non_snake_case)]
impl Auth {
    pub fn new(root_url: &str, credentials: Option<Credentials>) -> Result<Self, Error> {
        Ok(Self(Client::new(root_url, "auth", "v1", credentials)?))
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

    /// List Clients
    /// 
    /// Get a list of all clients.  With `prefix`, only clients for which
    /// it is a prefix of the clientId are returned.
    /// 
    /// By default this end-point will try to return up to 1000 clients in one
    /// request. But it **may return less, even none**.
    /// It may also return a `continuationToken` even though there are no more
    /// results. However, you can only be sure to have seen all results if you
    /// keep calling `listClients` with the last `continuationToken` until you
    /// get a result without a `continuationToken`.
    pub async fn listClients(&self, prefix: Option<&str>, continuationToken: Option<&str>, limit: Option<&str>) -> Result<Value, Error> {
        let method = "GET";
        let path = "clients/";
        let mut query = None;
        if let Some(q) = prefix {
            query.get_or_insert_with(Vec::new).push(("prefix", q));
        }
        if let Some(q) = continuationToken {
            query.get_or_insert_with(Vec::new).push(("continuationToken", q));
        }
        if let Some(q) = limit {
            query.get_or_insert_with(Vec::new).push(("limit", q));
        }
        let body = None;

        let resp = self.0.request(method, path, query, body).await?;

        Ok(resp.json().await?)
    }

    /// Get Client
    /// 
    /// Get information about a single client.
    pub async fn client(&self, clientId: &str) -> Result<Value, Error> {
        let method = "GET";
        let path = format!("clients/{}", urlencode(clientId));
        let query = None;
        let body = None;

        let resp = self.0.request(method, &path, query, body).await?;

        Ok(resp.json().await?)
    }

    /// Create Client
    /// 
    /// Create a new client and get the `accessToken` for this client.
    /// You should store the `accessToken` from this API call as there is no
    /// other way to retrieve it.
    /// 
    /// If you loose the `accessToken` you can call `resetAccessToken` to reset
    /// it, and a new `accessToken` will be returned, but you cannot retrieve the
    /// current `accessToken`.
    /// 
    /// If a client with the same `clientId` already exists this operation will
    /// fail. Use `updateClient` if you wish to update an existing client.
    /// 
    /// The caller's scopes must satisfy `scopes`.
    pub async fn createClient(&self, clientId: &str, payload: &Value) -> Result<Value, Error> {
        let method = "PUT";
        let path = format!("clients/{}", urlencode(clientId));
        let query = None;
        let body = Some(payload);

        let resp = self.0.request(method, &path, query, body).await?;

        Ok(resp.json().await?)
    }

    /// Reset `accessToken`
    /// 
    /// Reset a clients `accessToken`, this will revoke the existing
    /// `accessToken`, generate a new `accessToken` and return it from this
    /// call.
    /// 
    /// There is no way to retrieve an existing `accessToken`, so if you loose it
    /// you must reset the accessToken to acquire it again.
    pub async fn resetAccessToken(&self, clientId: &str) -> Result<Value, Error> {
        let method = "POST";
        let path = format!("clients/{}/reset", urlencode(clientId));
        let query = None;
        let body = None;

        let resp = self.0.request(method, &path, query, body).await?;

        Ok(resp.json().await?)
    }

    /// Update Client
    /// 
    /// Update an exisiting client. The `clientId` and `accessToken` cannot be
    /// updated, but `scopes` can be modified.  The caller's scopes must
    /// satisfy all scopes being added to the client in the update operation.
    /// If no scopes are given in the request, the client's scopes remain
    /// unchanged
    pub async fn updateClient(&self, clientId: &str, payload: &Value) -> Result<Value, Error> {
        let method = "POST";
        let path = format!("clients/{}", urlencode(clientId));
        let query = None;
        let body = Some(payload);

        let resp = self.0.request(method, &path, query, body).await?;

        Ok(resp.json().await?)
    }

    /// Enable Client
    /// 
    /// Enable a client that was disabled with `disableClient`.  If the client
    /// is already enabled, this does nothing.
    /// 
    /// This is typically used by identity providers to re-enable clients that
    /// had been disabled when the corresponding identity's scopes changed.
    pub async fn enableClient(&self, clientId: &str) -> Result<Value, Error> {
        let method = "POST";
        let path = format!("clients/{}/enable", urlencode(clientId));
        let query = None;
        let body = None;

        let resp = self.0.request(method, &path, query, body).await?;

        Ok(resp.json().await?)
    }

    /// Disable Client
    /// 
    /// Disable a client.  If the client is already disabled, this does nothing.
    /// 
    /// This is typically used by identity providers to disable clients when the
    /// corresponding identity's scopes no longer satisfy the client's scopes.
    pub async fn disableClient(&self, clientId: &str) -> Result<Value, Error> {
        let method = "POST";
        let path = format!("clients/{}/disable", urlencode(clientId));
        let query = None;
        let body = None;

        let resp = self.0.request(method, &path, query, body).await?;

        Ok(resp.json().await?)
    }

    /// Delete Client
    /// 
    /// Delete a client, please note that any roles related to this client must
    /// be deleted independently.
    pub async fn deleteClient(&self, clientId: &str) -> Result<(), Error> {
        let method = "DELETE";
        let path = format!("clients/{}", urlencode(clientId));
        let query = None;
        let body = None;

        let resp = self.0.request(method, &path, query, body).await?;

        resp.bytes().await?;
        Ok(())
    }

    /// List Roles (no pagination)
    /// 
    /// Get a list of all roles. Each role object also includes the list of
    /// scopes it expands to.  This always returns all roles in a single HTTP
    /// request.
    /// 
    /// To get paginated results, use `listRoles2`.
    pub async fn listRoles(&self) -> Result<Value, Error> {
        let method = "GET";
        let path = "roles/";
        let query = None;
        let body = None;

        let resp = self.0.request(method, path, query, body).await?;

        Ok(resp.json().await?)
    }

    /// List Roles
    /// 
    /// Get a list of all roles. Each role object also includes the list of
    /// scopes it expands to.  This is similar to `listRoles` but differs in the
    /// format of the response.
    /// 
    /// If no limit is given, all roles are returned. Since this
    /// list may become long, callers can use the `limit` and `continuationToken`
    /// query arguments to page through the responses.
    pub async fn listRoles2(&self, continuationToken: Option<&str>, limit: Option<&str>) -> Result<Value, Error> {
        let method = "GET";
        let path = "roles2/";
        let mut query = None;
        if let Some(q) = continuationToken {
            query.get_or_insert_with(Vec::new).push(("continuationToken", q));
        }
        if let Some(q) = limit {
            query.get_or_insert_with(Vec::new).push(("limit", q));
        }
        let body = None;

        let resp = self.0.request(method, path, query, body).await?;

        Ok(resp.json().await?)
    }

    /// List Role IDs
    /// 
    /// Get a list of all role IDs.
    /// 
    /// If no limit is given, the roleIds of all roles are returned. Since this
    /// list may become long, callers can use the `limit` and `continuationToken`
    /// query arguments to page through the responses.
    pub async fn listRoleIds(&self, continuationToken: Option<&str>, limit: Option<&str>) -> Result<Value, Error> {
        let method = "GET";
        let path = "roleids/";
        let mut query = None;
        if let Some(q) = continuationToken {
            query.get_or_insert_with(Vec::new).push(("continuationToken", q));
        }
        if let Some(q) = limit {
            query.get_or_insert_with(Vec::new).push(("limit", q));
        }
        let body = None;

        let resp = self.0.request(method, path, query, body).await?;

        Ok(resp.json().await?)
    }

    /// Get Role
    /// 
    /// Get information about a single role, including the set of scopes that the
    /// role expands to.
    pub async fn role(&self, roleId: &str) -> Result<Value, Error> {
        let method = "GET";
        let path = format!("roles/{}", urlencode(roleId));
        let query = None;
        let body = None;

        let resp = self.0.request(method, &path, query, body).await?;

        Ok(resp.json().await?)
    }

    /// Create Role
    /// 
    /// Create a new role.
    /// 
    /// The caller's scopes must satisfy the new role's scopes.
    /// 
    /// If there already exists a role with the same `roleId` this operation
    /// will fail. Use `updateRole` to modify an existing role.
    /// 
    /// Creation of a role that will generate an infinite expansion will result
    /// in an error response.
    pub async fn createRole(&self, roleId: &str, payload: &Value) -> Result<Value, Error> {
        let method = "PUT";
        let path = format!("roles/{}", urlencode(roleId));
        let query = None;
        let body = Some(payload);

        let resp = self.0.request(method, &path, query, body).await?;

        Ok(resp.json().await?)
    }

    /// Update Role
    /// 
    /// Update an existing role.
    /// 
    /// The caller's scopes must satisfy all of the new scopes being added, but
    /// need not satisfy all of the role's existing scopes.
    /// 
    /// An update of a role that will generate an infinite expansion will result
    /// in an error response.
    pub async fn updateRole(&self, roleId: &str, payload: &Value) -> Result<Value, Error> {
        let method = "POST";
        let path = format!("roles/{}", urlencode(roleId));
        let query = None;
        let body = Some(payload);

        let resp = self.0.request(method, &path, query, body).await?;

        Ok(resp.json().await?)
    }

    /// Delete Role
    /// 
    /// Delete a role. This operation will succeed regardless of whether or not
    /// the role exists.
    pub async fn deleteRole(&self, roleId: &str) -> Result<(), Error> {
        let method = "DELETE";
        let path = format!("roles/{}", urlencode(roleId));
        let query = None;
        let body = None;

        let resp = self.0.request(method, &path, query, body).await?;

        resp.bytes().await?;
        Ok(())
    }

    /// Expand Scopes
    /// 
    /// Return an expanded copy of the given scopeset, with scopes implied by any
    /// roles included.
    pub async fn expandScopes(&self, payload: &Value) -> Result<Value, Error> {
        let method = "POST";
        let path = "scopes/expand";
        let query = None;
        let body = Some(payload);

        let resp = self.0.request(method, path, query, body).await?;

        Ok(resp.json().await?)
    }

    /// Get Current Scopes
    /// 
    /// Return the expanded scopes available in the request, taking into account all sources
    /// of scopes and scope restrictions (temporary credentials, assumeScopes, client scopes,
    /// and roles).
    pub async fn currentScopes(&self) -> Result<Value, Error> {
        let method = "GET";
        let path = "scopes/current";
        let query = None;
        let body = None;

        let resp = self.0.request(method, path, query, body).await?;

        Ok(resp.json().await?)
    }

    /// Get Temporary Read/Write Credentials S3
    /// 
    /// Get temporary AWS credentials for `read-write` or `read-only` access to
    /// a given `bucket` and `prefix` within that bucket.
    /// The `level` parameter can be `read-write` or `read-only` and determines
    /// which type of credentials are returned. Please note that the `level`
    /// parameter is required in the scope guarding access.  The bucket name must
    /// not contain `.`, as recommended by Amazon.
    /// 
    /// This method can only allow access to a whitelisted set of buckets, as configured
    /// in the Taskcluster deployment
    /// 
    /// The credentials are set to expire after an hour, but this behavior is
    /// subject to change. Hence, you should always read the `expires` property
    /// from the response, if you intend to maintain active credentials in your
    /// application.
    /// 
    /// Please note that your `prefix` may not start with slash `/`. Such a prefix
    /// is allowed on S3, but we forbid it here to discourage bad behavior.
    /// 
    /// Also note that if your `prefix` doesn't end in a slash `/`, the STS
    /// credentials may allow access to unexpected keys, as S3 does not treat
    /// slashes specially.  For example, a prefix of `my-folder` will allow
    /// access to `my-folder/file.txt` as expected, but also to `my-folder.txt`,
    /// which may not be intended.
    /// 
    /// Finally, note that the `PutObjectAcl` call is not allowed.  Passing a canned
    /// ACL other than `private` to `PutObject` is treated as a `PutObjectAcl` call, and
    /// will result in an access-denied error from AWS.  This limitation is due to a
    /// security flaw in Amazon S3 which might otherwise allow indefinite access to
    /// uploaded objects.
    /// 
    /// **EC2 metadata compatibility**, if the querystring parameter
    /// `?format=iam-role-compat` is given, the response will be compatible
    /// with the JSON exposed by the EC2 metadata service. This aims to ease
    /// compatibility for libraries and tools built to auto-refresh credentials.
    /// For details on the format returned by EC2 metadata service see:
    /// [EC2 User Guide](http://docs.aws.amazon.com/AWSEC2/latest/UserGuide/iam-roles-for-amazon-ec2.html#instance-metadata-security-credentials).
    pub async fn awsS3Credentials(&self, level: &str, bucket: &str, prefix: &str, format: Option<&str>) -> Result<Value, Error> {
        let method = "GET";
        let path = format!("aws/s3/{}/{}/{}", urlencode(level), urlencode(bucket), urlencode(prefix));
        let mut query = None;
        if let Some(q) = format {
            query.get_or_insert_with(Vec::new).push(("format", q));
        }
        let body = None;

        let resp = self.0.request(method, &path, query, body).await?;

        Ok(resp.json().await?)
    }

    /// List Accounts Managed by Auth
    /// 
    /// Retrieve a list of all Azure accounts managed by Taskcluster Auth.
    pub async fn azureAccounts(&self) -> Result<Value, Error> {
        let method = "GET";
        let path = "azure/accounts";
        let query = None;
        let body = None;

        let resp = self.0.request(method, path, query, body).await?;

        Ok(resp.json().await?)
    }

    /// List Tables in an Account Managed by Auth
    /// 
    /// Retrieve a list of all tables in an account.
    pub async fn azureTables(&self, account: &str, continuationToken: Option<&str>) -> Result<Value, Error> {
        let method = "GET";
        let path = format!("azure/{}/tables", urlencode(account));
        let mut query = None;
        if let Some(q) = continuationToken {
            query.get_or_insert_with(Vec::new).push(("continuationToken", q));
        }
        let body = None;

        let resp = self.0.request(method, &path, query, body).await?;

        Ok(resp.json().await?)
    }

    /// Get Shared-Access-Signature for Azure Table
    /// 
    /// Get a shared access signature (SAS) string for use with a specific Azure
    /// Table Storage table.
    /// 
    /// The `level` parameter can be `read-write` or `read-only` and determines
    /// which type of credentials are returned.  If level is read-write, it will create the
    /// table if it doesn't already exist.
    pub async fn azureTableSAS(&self, account: &str, table: &str, level: &str) -> Result<Value, Error> {
        let method = "GET";
        let path = format!("azure/{}/table/{}/{}", urlencode(account), urlencode(table), urlencode(level));
        let query = None;
        let body = None;

        let resp = self.0.request(method, &path, query, body).await?;

        Ok(resp.json().await?)
    }

    /// List containers in an Account Managed by Auth
    /// 
    /// Retrieve a list of all containers in an account.
    pub async fn azureContainers(&self, account: &str, continuationToken: Option<&str>) -> Result<Value, Error> {
        let method = "GET";
        let path = format!("azure/{}/containers", urlencode(account));
        let mut query = None;
        if let Some(q) = continuationToken {
            query.get_or_insert_with(Vec::new).push(("continuationToken", q));
        }
        let body = None;

        let resp = self.0.request(method, &path, query, body).await?;

        Ok(resp.json().await?)
    }

    /// Get Shared-Access-Signature for Azure Container
    /// 
    /// Get a shared access signature (SAS) string for use with a specific Azure
    /// Blob Storage container.
    /// 
    /// The `level` parameter can be `read-write` or `read-only` and determines
    /// which type of credentials are returned.  If level is read-write, it will create the
    /// container if it doesn't already exist.
    pub async fn azureContainerSAS(&self, account: &str, container: &str, level: &str) -> Result<Value, Error> {
        let method = "GET";
        let path = format!("azure/{}/containers/{}/{}", urlencode(account), urlencode(container), urlencode(level));
        let query = None;
        let body = None;

        let resp = self.0.request(method, &path, query, body).await?;

        Ok(resp.json().await?)
    }

    /// Get DSN for Sentry Project
    /// 
    /// Get temporary DSN (access credentials) for a sentry project.
    /// The credentials returned can be used with any Sentry client for up to
    /// 24 hours, after which the credentials will be automatically disabled.
    /// 
    /// If the project doesn't exist it will be created, and assigned to the
    /// initial team configured for this component. Contact a Sentry admin
    /// to have the project transferred to a team you have access to if needed
    pub async fn sentryDSN(&self, project: &str) -> Result<Value, Error> {
        let method = "GET";
        let path = format!("sentry/{}/dsn", urlencode(project));
        let query = None;
        let body = None;

        let resp = self.0.request(method, &path, query, body).await?;

        Ok(resp.json().await?)
    }

    /// Get a client token for the Websocktunnel service
    /// 
    /// Get a temporary token suitable for use connecting to a
    /// [websocktunnel](https://github.com/taskcluster/taskcluster/tree/main/tools/websocktunnel) server.
    /// 
    /// The resulting token will only be accepted by servers with a matching audience
    /// value.  Reaching such a server is the callers responsibility.  In general,
    /// a server URL or set of URLs should be provided to the caller as configuration
    /// along with the audience value.
    /// 
    /// The token is valid for a limited time (on the scale of hours). Callers should
    /// refresh it before expiration.
    pub async fn websocktunnelToken(&self, wstAudience: &str, wstClient: &str) -> Result<Value, Error> {
        let method = "GET";
        let path = format!("websocktunnel/{}/{}", urlencode(wstAudience), urlencode(wstClient));
        let query = None;
        let body = None;

        let resp = self.0.request(method, &path, query, body).await?;

        Ok(resp.json().await?)
    }

    /// Get Temporary GCP Credentials
    /// 
    /// Get temporary GCP credentials for the given serviceAccount in the given project.
    /// 
    /// Only preconfigured projects and serviceAccounts are allowed, as defined in the
    /// deployment of the Taskcluster services.
    /// 
    /// The credentials are set to expire after an hour, but this behavior is
    /// subject to change. Hence, you should always read the `expires` property
    /// from the response, if you intend to maintain active credentials in your
    /// application.
    pub async fn gcpCredentials(&self, projectId: &str, serviceAccount: &str) -> Result<Value, Error> {
        let method = "GET";
        let path = format!("gcp/credentials/{}/{}", urlencode(projectId), urlencode(serviceAccount));
        let query = None;
        let body = None;

        let resp = self.0.request(method, &path, query, body).await?;

        Ok(resp.json().await?)
    }

    /// Authenticate Hawk Request
    /// 
    /// Validate the request signature given on input and return list of scopes
    /// that the authenticating client has.
    /// 
    /// This method is used by other services that wish rely on Taskcluster
    /// credentials for authentication. This way we can use Hawk without having
    /// the secret credentials leave this service.
    pub async fn authenticateHawk(&self, payload: &Value) -> Result<Value, Error> {
        let method = "POST";
        let path = "authenticate-hawk";
        let query = None;
        let body = Some(payload);

        let resp = self.0.request(method, path, query, body).await?;

        Ok(resp.json().await?)
    }

    /// Test Authentication
    /// 
    /// Utility method to test client implementations of Taskcluster
    /// authentication.
    /// 
    /// Rather than using real credentials, this endpoint accepts requests with
    /// clientId `tester` and accessToken `no-secret`. That client's scopes are
    /// based on `clientScopes` in the request body.
    /// 
    /// The request is validated, with any certificate, authorizedScopes, etc.
    /// applied, and the resulting scopes are checked against `requiredScopes`
    /// from the request body. On success, the response contains the clientId
    /// and scopes as seen by the API method.
    pub async fn testAuthenticate(&self, payload: &Value) -> Result<Value, Error> {
        let method = "POST";
        let path = "test-authenticate";
        let query = None;
        let body = Some(payload);

        let resp = self.0.request(method, path, query, body).await?;

        Ok(resp.json().await?)
    }

    /// Test Authentication (GET)
    /// 
    /// Utility method similar to `testAuthenticate`, but with the GET method,
    /// so it can be used with signed URLs (bewits).
    /// 
    /// Rather than using real credentials, this endpoint accepts requests with
    /// clientId `tester` and accessToken `no-secret`. That client's scopes are
    /// `['test:*', 'auth:create-client:test:*']`.  The call fails if the 
    /// `test:authenticate-get` scope is not available.
    /// 
    /// The request is validated, with any certificate, authorizedScopes, etc.
    /// applied, and the resulting scopes are checked, just like any API call.
    /// On success, the response contains the clientId and scopes as seen by
    /// the API method.
    /// 
    /// This method may later be extended to allow specification of client and
    /// required scopes via query arguments.
    pub async fn testAuthenticateGet(&self) -> Result<Value, Error> {
        let method = "GET";
        let path = "test-authenticate-get/";
        let query = None;
        let body = None;

        let resp = self.0.request(method, path, query, body).await?;

        Ok(resp.json().await?)
    }
}