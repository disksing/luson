# luson

luson is a self-hosted JSON API server.

## Examples

### Create

- Create JSON entry with content

```
curl -XPOST -H "Authorization:${KEY}" -i "http://${YOURHOST}/" -d \
  '{"app": "luson", "loveFrom": [{"language":"Go"}, {"editor":"vscode"}, "GitHub"]}'
# returns JSON entry id
201 Created
06e30e01-bed7-451b-b35b-48dee43f06d4
```

### Read

- Get JSON entry

```
curl -i "http://${YOURHOST}/${ID}?pretty"
# returns JSON entry
200 OK
{
  "app": "luson",
  "loveFrom": [
    {
      "language": "Go"
    },
    {
      "editor": "vscode"
    },
    "GitHub"
  ]
}
```
