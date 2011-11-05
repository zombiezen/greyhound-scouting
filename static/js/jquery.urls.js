/*
 *  jquery.urls.js
 *
 *  Created by Ross Light on 1/13/09.
*/

var _jquery_urls_gRootURL = null;
var _jquery_urls_gStaticURL = null;

function _jquery_urls_append_url(root, child)
{
    if (child.charAt(0) == "/")
    {
        child = child.substring(1, child.length);
    }
    return root + child;
}

function _jquery_urls_normalize_root(url)
{
    if (url.charAt(url.length - 1) != "/")
    {
        url += "/";
    }
    return url;
}

jQuery.siteURL = function(url)
{
    return _jquery_urls_append_url(_jquery_urls_gRootURL, url);
};

jQuery.setSiteRoot = function(url)
{
    _jquery_urls_gRootURL = _jquery_urls_normalize_root(url);
    if (_jquery_urls_gStaticURL == null)
    {
        _jquery_urls_gStaticURL = _jquery_urls_normalize_root(
            this.siteURL("/static/"));
    }
};

jQuery.staticURL = function(url)
{
    return _jquery_urls_append_url(_jquery_urls_gStaticURL, url);
};

jQuery.setStaticRoot = function(url)
{
    _jquery_urls_gStaticURL = _jquery_urls_normalize_root(url);
};
