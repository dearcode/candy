package libcore.io;

import com.android.volley.AuthFailureError;
import com.android.volley.NetworkResponse;
import com.android.volley.ParseError;
import com.android.volley.Response;
import com.android.volley.Response.ErrorListener;
import com.android.volley.Response.Listener;
import com.android.volley.toolbox.HttpHeaderParser;
import com.forlong401.log.transaction.utils.LogUtils;
import com.google.gson.Gson;
import com.google.gson.GsonBuilder;
import com.google.gson.JsonSyntaxException;

import net.dearcode.candy.controller.CustomeApplication;
import net.dearcode.candy.util.LogUtil;

import java.io.UnsupportedEncodingException;
import java.util.HashMap;
import java.util.Map;

import static com.forlong401.log.transaction.log.manager.LogManager.getManager;

/**
 * 返回json格式
 * @author lxq_x
 *
 * @param <T>
 */
public class GsonRequest<T> extends BaseRequst<T> {
    private final Listener<T> mListener;  
    private Gson mGson;  
    private Class<T> mClass;
    private Map<String, String> mParams;
    private String mUrl;

    public GsonRequest(int method, Map<String, String> params, String url, Class<T> clazz, Listener<T> listener,
                       ErrorListener errorListener) {
        super(method, url, errorListener);  
        mGson = new GsonBuilder().enableComplexMapKeySerialization().create(); 
        mClass = clazz;  
        mListener = listener; 
        mParams = params;
        mUrl = url;
    }  
  
    public GsonRequest(String url, Map<String, String> params, Class<T> clazz, Listener<T> listener,
                       ErrorListener errorListener) {
        this(Method.GET, params, url, clazz, listener, errorListener);
    }  
    
    @Override
    public Map<String, String> getHeaders() throws AuthFailureError {
        Map<String, String> headers = super.getHeaders();
        //API 网关相关
        //String appKey = "23337485";
        //String appSecret = "95dc584d68da2853d7006a6997a152cd";
        //headers.put("Accept", "application/json");
        /*try {
            MessageDigest md5 = MessageDigest.getInstance("MD5");
            headers.put("Content-MD5", StringUtils.newStringUtf8(Base64.encodeBase64(md5.digest(getBody()), false)));
        } catch (NoSuchAlgorithmException e) {
            e.printStackTrace();
        }*/
        //headers.put("Content-Type", "application/octet-stream");
       /* try {
            headers = HttpRequestSignUtil.initialBasicHeader(headers, appKey, appSecret, "post", mUrl, null, null);
            return headers;
        } catch (MalformedURLException e) {
            e.printStackTrace();
        }*/
        return headers;
    }
    
	@Override
	protected Map<String, String> getParams() throws AuthFailureError {
		Map<String, String> params = super.getParams();
		if(params == null){
			params = new HashMap<String, String>();
		}
		params.putAll(mParams);
		return params;
	}
  
    @Override
    protected Response<T> parseNetworkResponse(NetworkResponse response) {  
    	super.parseNetworkResponse(response);
        String jsonString = null;
        try {  
            jsonString = new String(response.data,
			         HttpHeaderParser.parseCharset(response.headers));
            LogUtil.d("GSON", "[" +  mParams.get("action") + " ]---返回的GSON数据是  " + jsonString);
            return Response.success(mGson.fromJson(jsonString, mClass),
                    HttpHeaderParser.parseCacheHeaders(response));
        } catch (UnsupportedEncodingException e) {
            return Response.error(new ParseError(e));  
        } catch (JsonSyntaxException e){
            getManager(CustomeApplication.getInstance()).log("GSON_Error",
                    "Json解析异常[ACTION=" +  mParams.get("action") + "]__返回数据：" + jsonString,
                    LogUtils.LOG_TYPE_2_FILE_AND_LOGCAT);
            return Response.error(new ParseError(e));
        }
    }  
  
    @Override
    protected void deliverResponse(T response) {  
        mListener.onResponse(response);  
    }  
}
