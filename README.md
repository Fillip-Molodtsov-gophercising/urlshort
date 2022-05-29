# URL shortener

The goal of this exercise is to create an [http.Handler](https://golang.org/pkg/net/http/#Handler) 
that will look at the path of any incoming web request and determine 
if it should redirect the user to a new page, much like URL shortener would.

There is a POST `/create` endpoint which you could call for a redirect pair to be stored in BoltDB.

There is a possibility to input an array of routing redirects from a yaml file via a flag.
They also will be stored in BoltDB.

For instance, if we have a redirect setup for `/dt` to `https://www.deepl.com/translator#` 
we would look for any incoming web requests with the path `/dt` and redirect them.