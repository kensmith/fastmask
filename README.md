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

Then, on Linux, (other OSes left as an exercise)

$ mkdir -p $HOME/.config/fastmask
$ chmod 700 $HOME/.config/fastmask
$ cat << EOF > $HOME/.config/fastmask/config.json
{
  token: "<your fastmail API token>",
}
EOF
$ chmod 600 $HOME/.config/fastmask/config.json

vim:tw=60:
