
# werf.io site backend

HTTP-server to make some logic for website (werf.io).

## GET `/api/status`

Get status info (JSON).

- 'status' — `ok` or `error`.
- 'msg' — Empty if `status` is `ok`, otherwise contains text representation of the error.
- `rootVersion` — Version to show as main. E.g. - `v1.2.4+fix18`.
- `rootVersionURL` — URL location for RootVersion. E.g. - `v1.2.4-plus-fix18`.
- `multiwerf` — content of the used multiwerf.json (info about which versions belong to update channels)

Example:
```json
{
  "status": "ok",
  "msg": "",
  "rootVersion": "v1.2.4+fix18",
  "rootVersionURL": "v1.2.4-plus-fix18",
  "multiwerf": [
    {
      "group": "1.0",
      "channels": [
        {
          "name": "alpha",
          "version": "v1.0.13+fix5"
        },
        {
          "name": "beta",
          "version": "v1.0.13+fix5"
        },
        {
          "name": "ea",
          "version": "v1.0.13+fix4"
        },
        {
          "name": "stable",
          "version": "v1.0.13+fix4"
        },
        {
          "name": "rock-solid",
          "version": "v1.0.13+fix4"
        }
      ]
    },
    {
      "group": "1.1",
      "channels": [
        {
          "name": "alpha",
          "version": "v1.1.23+fix15"
        },
        {
          "name": "beta",
          "version": "v1.1.23+fix14"
        },
        {
          "name": "ea",
          "version": "v1.1.23+fix14"
        },
        {
          "name": "stable",
          "version": "v1.1.23+fix6"
        },
        {
          "name": "rock-solid",
          "version": "v1.1.22+fix37"
        }
      ]
    },
    {
      "group": "1.2",
      "channels": [
        {
          "name": "alpha",
          "version": "v1.2.5+fix10"
        },
        {
          "name": "beta",
          "version": "v1.2.5+fix8"
        },
        {
          "name": "ea",
          "version": "v1.2.4+fix18"
        }
      ]
    }
  ]
}
```