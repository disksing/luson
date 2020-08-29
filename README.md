# luson

luson is a self-hosted JSON API server.

## Examples

NOTE: Remember to replace following `${YOURHOST}` with real host (usually `localhost:42195`), `${KEY}`s with real API key, and `${ID}`s with real JSON id.

### Create

- Create JSON entry with content

```
curl -XPOST -H "Authorization:${KEY}" -i http://${YOURHOST}/ -d '{"hello": "luson"}'
# returns JSON entry id
201 Created
06e30e01-bed7-451b-b35b-48dee43f06d4
```

- Create empty JSON entry

```
curl -XPOST -H "Authorization:${KEY}" -i http://${YOURHOST}/
# returns JSON entry id
201 Created
101a53e1-3e49-4cc8-b771-0ff55f66c72a
```

