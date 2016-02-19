-- @Author: detailyang
-- @Date:   2016-02-19 17:47:58
-- @Last Modified by:   detailyang
-- @Last Modified time: 2016-02-19 19:19:28

function request()
    req.header.who = "xiaoyu"
    req.header["x-kdt-scheme"] = req.url.scheme
    if req.url.scheme == "https" then
        req.url.scheme = "http"
    end
    req.url.host = req.header.host
end