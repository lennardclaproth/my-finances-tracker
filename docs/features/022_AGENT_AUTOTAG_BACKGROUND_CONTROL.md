## [022] Agent Auto-Tag Background Control

### Summary
Agent-based cashflow auto-tagging can be explicitly enabled or disabled through configuration.  
When disabled, the agent tagger background worker is not registered and therefore does not run.

### Why This Feature Exists
Environments may need to run the API without external agent dependencies (for example local development, incident mitigation, or controlled rollouts).  
A first-class toggle prevents partial startup behavior and makes the background job topology explicit.

### Configuration Contract
- `agent.enabled: true|false` controls whether the tagger job is registered.
- `agent.agent_base_url` and `agent.default_tag_agent_id` are required only when `agent.enabled` is `true`.
- `agent.default_tag_agent_id` must be a valid UUID when enabled.

Example:

```yaml
agent:
  enabled: true
  agent_base_url: http://localhost:8001/api
  default_tag_agent_id: "4cf3c137-4228-44fe-8f56-cd8ed83a8103"
```

### Runtime Behavior
- `agent.enabled=true`
  - `TaggerJob` is created and started by the job manager.
  - Untagged cashflow transactions are processed via the agent runner flow.
- `agent.enabled=false`
  - `TaggerJob` is not registered with the manager.
  - No agent auto-tag background polling/processing occurs.

### Error Handling and Logging
- Startup logs explicitly indicate whether agent auto-tagging is enabled or disabled.
- If an invalid UUID bypasses config validation and reaches composition, startup keeps running and logs an error while skipping tagger registration.

### Code References
- Configuration model and validation:
  - [config.go](C:/personal/git/my-finances-tracker/apps/api/internal/config/config.go)
- Job composition toggle:
  - [main.go](C:/personal/git/my-finances-tracker/apps/api/cmd/server/main.go)
- Managed job set:
  - [manager.go](C:/personal/git/my-finances-tracker/apps/api/internal/jobs/manager.go)
  - [tagger.go](C:/personal/git/my-finances-tracker/apps/api/internal/jobs/tagger.go)

### Validation Coverage
- Config validation coverage for enabled/disabled agent settings:
  - [config_test.go](C:/personal/git/my-finances-tracker/apps/api/internal/config/config_test.go)
- Startup composition coverage for including/excluding `TaggerJob`:
  - [main_jobs_test.go](C:/personal/git/my-finances-tracker/apps/api/cmd/server/main_jobs_test.go)
