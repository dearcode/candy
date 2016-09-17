package net.dearcode.candy.controller;

import android.app.Application;
import android.content.Context;

import com.forlong401.log.transaction.log.manager.LogManager;

/**
 * Created by 水寒 on 2016/9/17.
 * 自定义application
 */

public class CustomeApplication extends Application {

    private static CustomeApplication mInstance;

    public CustomeApplication(){
        mInstance = this;
    }

    @Override
    public void onCreate() {
        super.onCreate();
        LogManager.getManager(this).registerCrashHandler();
    }

    @Override
    public void onTerminate() {
        super.onTerminate();
        LogManager.getManager(this).unregisterCrashHandler();
    }

    public Context getContext(){
        return mInstance.getContext();
    }

    public static CustomeApplication getInstance(){
        return mInstance;
    }
}
