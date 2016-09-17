package net.dearcode.candy.manage.config;

/**
 * Created by 水寒 on 2016/9/17.
 * 系统相关配置管理
 */

public class SystemConfigManage extends BaseConfigManage {

    public static final String CONFIG_NAME = "system_config";
    private static SystemConfigManage mInstance;

    private static final String KEY_CURRENT_IP_ADDRESS = "currentIpAddress";

    private String currentIpAddress;      //当前ip

    protected SystemConfigManage() {
        super(CONFIG_NAME);
        currentIpAddress = mSharePreference.getString(KEY_CURRENT_IP_ADDRESS, null);
    }

    public static SystemConfigManage getInstance(){
        if(mInstance == null){
            mInstance = new SystemConfigManage();
        }
        return mInstance;
    }

    public String getCurrentIpAddress(){
        return currentIpAddress;
    }

    public void setCurrentIpAddress(String ipAddress){
        currentIpAddress = ipAddress;
        setConfig(KEY_CURRENT_IP_ADDRESS, currentIpAddress);
    }
}
