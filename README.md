# fastmask

Minimal, low dependency Javascript program to create
Fastmail masked emails using the JMAP API.

## Setup:

Generate your API token from your Fastmail account:
* Settings -> Privacy & Security -> Integrations -> API
  tokens
  * You only need to grant the "Masked Email" scope.
  * Don't set it to "Read-only access." Leave that
    unchecked.
  * https://app.fastmail.com/settings/security/tokens

Then, on Linux, (other OSes left as an exercise)

```
$ mkdir -p $HOME/.config/fastmask
$ chmod 700 $HOME/.config/fastmask
$ cat << EOF > $HOME/.config/fastmask/config.json
{
  "token": "<your fastmail API token>",
}
EOF
$ chmod 600 $HOME/.config/fastmask/config.json
```

You can optionally set `"prefix"` in `config.json` to cause
your generated masked emails to have a constant prefix as
opposed to a randomly generated one.

## Usage:

General:

```
$ fastmask <domain>
```

Example:

```
$ fastmask example.com
```

This emits JSON with the following form:

```
{
  "prefix": "mshqh",
  "domain": "example.com",
  "email": "mshqh.xxxxx@fastmail.com"
}
```

Which you can deal with programmatically with something like this:

```
$ fastmasil example.com | jq '.email'
```

## Related

This project was inspired by
https://github.com/dvcrn/maskedemail-cli and
https://github.com/fastmail/JMAP-Samples. It is strictly
targeted at creating new masked emails. None of the other
APIs are covered in order to be as simple as possible.

## Installation

Just `chmod 755` the program and copy it to your
`$HOME/.local/bin` or where ever you keep your personal
binaries. You'll also need [Node.js](https://nodejs.org) for
your platform.

`License: MIT`

vim:tw=60:
