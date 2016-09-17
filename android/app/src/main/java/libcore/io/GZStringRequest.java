package libcore.io;

import com.android.volley.AuthFailureError;
import com.android.volley.Response.ErrorListener;
import com.android.volley.Response.Listener;
import com.android.volley.toolbox.StringRequest;

import java.util.HashMap;
import java.util.Map;

public class GZStringRequest extends StringRequest{
	private Map<String, String> mParams;
	public GZStringRequest(Map<String, String> params, String url, Listener<String> listener,
						   ErrorListener errorListener) {
		super(url, listener, errorListener);
		mParams = params;
	}

	public GZStringRequest(Map<String, String> params, int method, String url, Listener<String> listener,
						   ErrorListener errorListener) {
		super(method, url, listener, errorListener);
		mParams = params;
	}

	@Override
	protected Map<String, String> getParams() throws AuthFailureError {
		Map<String, String> params = super.getParams();
		if(params == null){
			params = new HashMap<String, String>();
		}
		params.put("Content-Type", "application/x-www-form-urlencoded;charset=utf-8");
		params.putAll(mParams);
		return params;
	}
}
