. ~/.profile-functions

color_my_prompt

GOPATH=$(go env GOPATH)
export PATH

PATH="/usr/local/bin:$GOPATH/bin:$PATH"
export PATH
