package net.dearcode.candy.controller.base;

import android.annotation.SuppressLint;
import android.app.Activity;
import android.os.Message;

import com.forlong401.log.transaction.log.manager.LogManager;

/**
 * 基础控制器
 * @author lxq_x
 *
 */
public class BaseController {

	protected boolean mIsTop;

	public void onCreate(Activity context){
		//EventBus.getDefault().register(context);
		//注册收集日志
		LogManager.getManager(context.getApplicationContext()).registerActivity(context);
	}

	public void onDestroy(Activity context) {
		//EventBus.getDefault().unregister(context);
		//注销收集日志
		LogManager.getManager(context.getApplicationContext()).unregisterActivity(context);
	}

	public void onResume(Activity context){

	}
	
	public void onPause(Activity context){

	}

	@SuppressLint("NewApi")
	public void onEventMainThread(final Activity context, Message message){
		//BaseEvent.handleEvent(context, message);
	}
}
