package net.dearcode.candy.manage.config;

import android.content.Context;
import android.content.SharedPreferences;

import net.dearcode.candy.controller.CustomeApplication;

import java.util.Set;

/**
 * Created by 水寒 on 2016/9/17.
 * 配置文件管理类基类
 */

public class BaseConfigManage {
    protected SharedPreferences mSharePreference;

    protected BaseConfigManage(String shareName){
        mSharePreference = CustomeApplication.getInstance().getContext()
                .getSharedPreferences(shareName, Context.MODE_PRIVATE);
    }

    /**
     * 设置配置
     * @param key
     * @param value
     */
    protected void setConfig(String key, int value){
        SharedPreferences.Editor editor = mSharePreference.edit();
        editor.putInt(key, value);
        editor.apply();
    }

    /**
     * 设置配置
     * @param key
     * @param value
     */
    protected void setConfig(String key, String value){
        SharedPreferences.Editor editor = mSharePreference.edit();
        editor.putString(key, value);
        editor.apply();
    }

    /**
     * 设置配置
     * @param key
     * @param value
     */
    protected void setConfig(String key, long value){
        SharedPreferences.Editor editor = mSharePreference.edit();
        editor.putLong(key, value);
        editor.apply();
    }

    /**
     * 设置配置
     * @param key
     * @param value
     */
    protected void setConfig(String key, Set<String> value){
        SharedPreferences.Editor editor = mSharePreference.edit();
        editor.putStringSet(key, value);
        editor.apply();
    }

    /**
     * 设置配置
     * @param key
     * @param value
     */
    protected void setConfig(String key, boolean value){
        SharedPreferences.Editor editor = mSharePreference.edit();
        editor.putBoolean(key, value);
        editor.apply();
    }
}
