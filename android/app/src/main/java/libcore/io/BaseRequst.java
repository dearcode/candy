package libcore.io;

import com.android.volley.AuthFailureError;
import com.android.volley.NetworkResponse;
import com.android.volley.Request;
import com.android.volley.Response;
import com.android.volley.Response.ErrorListener;

import java.util.HashMap;
import java.util.Map;

public class BaseRequst<T> extends Request<T>{
	
    private String mHeader;
    public String cookieFromResponse;


	public BaseRequst(int method, String url, ErrorListener listener) {
		super(method, url, listener);
	}

	@Override
	protected void deliverResponse(T arg0) {
		// TODO Auto-generated method stub
		
	}
	
	@Override
	public Map<String, String> getHeaders() throws AuthFailureError {
		Map<String, String> headers = new HashMap<String, String>();
		//headers.put("Cookie", NetworkConfigManage.getInstance().getCookie());
		//headers.put(SystemHeader.SIGN_PARAM, HttpRequestSignUtil.getEncodeInput());
		return headers;
	}

	@Override
	protected Response<T> parseNetworkResponse(NetworkResponse response) {
		/*mHeader = response.headers.toString();
		//使用正则表达式从reponse的头中提取cookie内容的子串
		Pattern pattern = Pattern.compile("Set-Cookie.*?;");
		Matcher m = pattern.matcher(mHeader);
		if(m.find()){
			cookieFromResponse = m.group();
			//LogUtil.d("cookie from server "+ cookieFromResponse);
		}
		//去掉cookie末尾的分号
		if(cookieFromResponse != null){
			cookieFromResponse = cookieFromResponse.substring(11,cookieFromResponse.length()-1);
			if(cookieFromResponse.contains("PHPSESSID")){
				NetworkConfigManage.getInstance().setCookie(cookieFromResponse);
			}
		}*/
		return null;
	}
}
