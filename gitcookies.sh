touch ~/.gitcookies
chmod 0600 ~/.gitcookies

git config --global http.cookiefile ~/.gitcookies

tr , \\t <<\__END__ >>~/.gitcookies
go.googlesource.com,FALSE,/,TRUE,2147483647,o,git-joseph.giantswarm.io=1/vxlFq4LObBrj22Iiy5AKT3g206S0oYp-lMhS73s673Q
go-review.googlesource.com,FALSE,/,TRUE,2147483647,o,git-joseph.giantswarm.io=1/vxlFq4LObBrj22Iiy5AKT3g206S0oYp-lMhS73s673Q
__END__
