-- @Author: detailyang
-- @Date:   2016-02-15 19:58:06
-- @Last Modified by:   detailyang
-- @Last Modified time: 2016-02-19 16:21:47

function request()
    req.header.copyright = "hijack"
    req.header.who = "xiaoyu"
    req.header["x-kdt-scheme"] = req.url.scheme
    if req.url.scheme == "https" then
        req.url.scheme = "http"
    end
    req.url.host = req.header.host
end
