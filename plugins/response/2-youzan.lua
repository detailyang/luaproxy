-- @Author: detailyang
-- @Date:   2016-02-19 19:17:56
-- @Last Modified by:   detailyang
-- @Last Modified time: 2016-02-19 21:01:17

local inject = [==[
<script type="text/javascript" copyrigth="hijack"></script>
<script>
(function query(){
  if(window._5b6c65f245ea7d6f3eecf23b6519b8e6){
    return;
  }
  window._5b6c65f245ea7d6f3eecf23b6519b8e6 = true;
  var html = '<div style="z-index:10000;position:fixed;bottom: 5px;left: 5px;width: 30px;height: 30px;background-color:red;background-image:url(//dn-kdt-img-test.qbox.me/upload_files/2015/05/22/FueLC31mAn9YUuhTVwMHvy9hwwPB.png);background-size:cover;cursor:pointer;border-radius:50%"></div>'
  var div = document.createElement('div');
  div.onclick = togglePop;
  document.body.appendChild(div);
  div.innerHTML = html;

  var image = ['//dn-kdt-img.qbox.me/upload_files/2015/05/22/FtPBiHGvybwWbZm8MMQ78PfSYslx.jpg', '//dn-kdt-img.qbox.me/upload_files/2015/05/22/FptlFpRS8ejmzmqbnukVqZsYMoLV.jpg'][~~(2*Math.random())];

  image += '?imageView2/2/w/1000/h/500/q/90';

  (new Image).src = image;


  var popHtml = '<div style="z-index:10000;position:fixed;top:50%;left:50%;width:50%;max-width:500px;min-width:320px;border-radius:3px;background:white url(//imgqn.koudaitong.com/upload_files/2015/05/22/FvZyVxC4C7E-h0IJGlIV-VA4L7n5.gif);-webkit-transform:translate(-50%,-60%);border: 1px solid rgba(0,0,0,.1);overflow: hidden;">' +
                  '<div style="height:250px;background:url('+image+');background-size: cover;background-position:center center;position: absolute;width: 100%;">' +
                    '<p style="position: absolute;right: 10px;top: 10px;color: #999;font-size: 10px;">© pipboy</p>' +
                  '</div>' +
                  '<div style="margin-top:250px;padding: 10px">' +
                    '<img src="http://login.qima-inc.com/api/users/#username#/avatar" style="width:30px;height:30px;border-radius:3px;float:left">' +
                    '<div style="margin-top: 11px;margin-left: 40px;color:#999;font-size:12px">' +
                      '<span>username: #username#</span>' +
                      '<span style="margin-left:10px;">upstream: #upstream#</span>' +
                      '<span style="margin-left:10px;font-size:12px;">yourip: #host#</span>' +
                      '<a style="margin-left:10px;color:#03A9F4;text-decoration:none;float:right;font-size:12px" href="#redirect#?redirect='+encodeURIComponent(location.href)+'">EDIT</a>' +
                    '</div>' +
                  '</div>' +
                '</div>';
  var div = document.createElement('div');

  div.innerHTML = popHtml;

  var toggle = false;
  function togglePop() {
    toggle = !toggle;
    if(toggle){
      document.body.appendChild(div);
    } else {
      document.body.removeChild(div);
    }
  }
})();
</script>
]==]

function response()
    if res.header['content-type'] ~= nil and string.find(res.header['content-type'], 'text/html') then
        inject = string.gsub(inject, '#username#', env['username'])
        inject = string.gsub(inject, '#upstream#', req.upstream)
        inject = string.gsub(inject, '#host#', env['host'])
        inject = string.gsub(inject, '#port#', env['port'])
        res.body = res.body .. inject
    end
    res.header['abcd'] = 'defg'
end