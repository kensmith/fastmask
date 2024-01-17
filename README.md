# fastmask

Minimal, low dependency program to create Fastmail masked
emails using their JMAP API.

Usage:

Generate your API token from your Fastmail account:
* Settings -> Privacy & Security -> Integrations -> API
  tokens
  * You only need to grant the "Masked Email" scope.
  * Don't set it to "Read-only access." Leave that
    unchecked.
  * https://app.fastmail.com/settings/security/tokens

Then, on Linux, (other OSes left as an exercise)

$ mkdir -p $HOME/.config/fastmask
$ chmod 700 $HOME/.config/fastmask
$ cat << EOF > $HOME/.config/fastmask/config.json
{
  token: "<your fastmail API token>",
}
EOF
$ chmod 600 $HOME/.config/fastmask/config.json

This project was inspired by
https://github.com/dvcrn/maskedemail-cli and
https://github.com/fastmail/JMAP-Samples. It is strictly
targeted at creating new masked emails. None of the other
APIs are covered in order to be as simple as possible.

License: MIT

vim:tw=60:
