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

- Get partial JSON entry

```
curl -i "http://${YOURHOST}/${ID}/loveFrom/1/editor"
# returns partial JSON
200 OK
"vscode"
```

### Update

- Update full JSON entry

```
curl -XPUT -H "Authorization:${KEY}" -i "http://${YOURHOST}/${ID}" -d \
  '{"app": "luson", "version": "v0.1", "loveFrom": [{"language":"Go"}, {"editor":"vscode"}, "GitHub"]}'

200 OK
```

- Update partial JSON entry

```
curl -XPUT -H "Authorization:${KEY}" -i "http://${YOURHOST}/${ID}/loveFrom/language" -d \
  '["Go", "markdown"]'

200 OK
```

