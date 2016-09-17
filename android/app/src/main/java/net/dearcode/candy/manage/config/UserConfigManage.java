package net.dearcode.candy.manage.config;

/**
 * Created by 水寒 on 2016/9/17.
 * 用户相关配置管理
 */

public class UserConfigManage extends BaseConfigManage {

    public static final String CONFIG_NAME = "user_config";
    private static UserConfigManage mInstance;

    public static final String KEY_USER_ID = "userId";

    public String userId;

    protected UserConfigManage() {
        super(CONFIG_NAME);
        userId = mSharePreference.getString(KEY_USER_ID, null);
    }

    public static UserConfigManage getInstance(){
        if(mInstance == null){
            mInstance = new UserConfigManage();
        }
        return mInstance;
    }

    public String getUserId() {
        return userId;
    }

    public void setUserId(String userId) {
        this.userId = userId;
        setConfig(KEY_USER_ID, userId);
    }
}
