package libcore.io;

import net.dearcode.candy.manage.config.SystemConfigManage;
import net.dearcode.candy.manage.config.UserConfigManage;

import java.io.UnsupportedEncodingException;
import java.util.Arrays;
import java.util.Collection;
import java.util.HashMap;
import java.util.HashSet;
import java.util.Iterator;
import java.util.Map;
import java.util.Map.Entry;
import java.util.Set;

/**
 * 请求参数
 * @author PeggyTong
 *
 */
public class RequestParam {
	public static final String REAL_BASE_URL = "xxxx";    							//正式服地址
	public static final String TEST_BASE_URL = "https://candy.dearcode.net:9000"; 							//测试服地址

	//public static String BASE_URL = SystemConfigManage.getInstance().getCurrentIpAddress();  			//正式发布
	public static String BASE_URL = TEST_BASE_URL;
	public enum URLType{
		userUrl,
		bbsUrl,
		cvUrl,
		catUrl
	}

	/**
	 * 服务器路径
	 */
	//用户
	public static String USER_URL = BASE_URL + "/interface/interface.php";
	//论坛
	public static String BBS_URL = BASE_URL + "/interface/discuz.php";
	//声优
	public static String CV_URL = BASE_URL + "/interface/interface.php";
	//产品
	public static String CAT_URL = BASE_URL + "/interface/shop.php ";
	
	/**
	 * 更换当前ip
	 * @return
	 */
	public static String switchIpAddress(String customUrl){
		if(customUrl == null){
			if((REAL_BASE_URL).equals(SystemConfigManage.getInstance().getCurrentIpAddress())){
				BASE_URL = TEST_BASE_URL;
			}else{
				BASE_URL = REAL_BASE_URL;
			}
		}else{
			BASE_URL = customUrl;
		}
		USER_URL = BASE_URL + "/interface.php";
		BBS_URL = BASE_URL + "/discuz.php";
		CV_URL = BASE_URL + "/interface.php";
		CAT_URL = BASE_URL + "/shop.php";
		SystemConfigManage.getInstance().setCurrentIpAddress(RequestParam.BASE_URL);
		return RequestParam.BASE_URL;
	}


	private Map<String, String> mParams = new HashMap<String, String>();

	private String mUrl;
	private String mParamsPath;

	public RequestParam(URLType urlType){
		switch (urlType) {
		case userUrl:
			mUrl = USER_URL;
			break;
		case bbsUrl:
			mUrl = BBS_URL;
			break;
		case cvUrl:
			mUrl = CV_URL;
			break;
		case catUrl:
			mUrl = CAT_URL;
			break;
		default:
			break;
		}
	}

	/**
	 * 设置参数
	 * @param key
	 * @param value
	 */
	public RequestParam setParams(String key, String value){
		//遇到空格需要转换
		//if(value != null){
		//	value = value.trim().replaceAll(" ", "_");
		//}
		if(value == null) value = "";
		mParams.put(key, value);
		return this;
	}

	public RequestParam setParams(String key, double value){
		mParams.put(key, String.valueOf(value));
		return this;
	}

	public RequestParam setParams(String key, long value){
		mParams.put(key, String.valueOf(value));
		return this;
	}

	/**
	 * 设置参数
	 * @param key
	 * @param value
	 * @return
	 */
	public RequestParam setParams(String key, int value){
		mParams.put(key, String.valueOf(value));
		return this;
	}

	/**
	 * 设置参数
	 * @param key
	 * @param values
	 */
	public RequestParam setParams(String key, Collection<String> values){
		if(values == null || values.isEmpty()) {
			setParams(key, "");
			return this;
		}
		Iterator<String> iterator = values.iterator();
		StringBuffer value = new StringBuffer();
		while(iterator.hasNext()){
			value.append(iterator.next());
			value.append(",");
		}
		String valueStr = value.substring(0, value.lastIndexOf(","));
		setParams(key, valueStr);
		return this;
	}

	/**
	 * 设置参数
	 * @param key
	 * @param values
	 * @return
	 */
	public RequestParam setParams(String key, String[] values){
		if(values == null || values.length == 0){
			setParams(key, "");
			return this;
		}
		Set<String> set = new HashSet<String>(Arrays.asList(values));
		setParams(key, set);
		return this;
	}

	/**
	 * 设置参数
	 * @param key
	 * @param values
	 * @return
	 */
	public RequestParam setParams(String key, int[] values){
		StringBuffer value = new StringBuffer();
		for(int i = 0; i < values.length; i++){
			value.append(values[i]);
			value.append(",");
		}
		String valueStr = value.substring(0, value.lastIndexOf(","));
		setParams(key, valueStr);
		return this;
	}

	/**
	 * 获取参数Map
	 * @return
	 */
	public Map<String, String> getParams(){
		if(!mParams.containsKey("uid")){
			setParams("uid", UserConfigManage.getInstance().getUserId());
		}
		return mParams;
	}

	@Override
	public String toString() {
		//生成字符串
		mParamsPath = getRquestUrl() + generatedAddress(mParams);
		String resultPath = "";
		try {
			resultPath = new String(mParamsPath.getBytes(), "ISO-8859-1");
		} catch (UnsupportedEncodingException e) {
			e.printStackTrace();
		}
		return resultPath;
	}

	/**
	 * 获取请求的地址链接
	 */
	public String getRquestUrl(){
		return mUrl;
	}

	/**
	 * 生成字符串地址参数
	 * @param params
	 * @return
	 */
	public static String generatedAddress(Map<String, String> params){
		if(params == null || params.isEmpty()) return "";
		StringBuffer sbuffer = new StringBuffer("?");
		Set<Entry<String, String>> set = params.entrySet();
		Iterator<Entry<String, String>> iterator = set.iterator();
		iterator.hasNext();
		Entry<String, String> entry = iterator.next();
		sbuffer.append(entry.getKey()).append("=").append(entry.getValue());
		while(iterator.hasNext()){
			entry = iterator.next();
			sbuffer.append("&&").append(entry.getKey()).append("=").append(entry.getValue());
		}
		return sbuffer.toString();
	}
}
