package net.dearcode.candy.controller.base;

import android.app.Activity;
import android.os.Bundle;
import android.os.Message;

/**
 * 所有Activity的基类
 * @author lxq_x
 *
 */
public class BaseActivity extends Activity {
	
	private BaseController mBaseController;
	
	public BaseActivity() {
		super();
		mBaseController = new BaseController();
	}

	@Override
	protected void onCreate(Bundle savedInstanceState) {
		super.onCreate(savedInstanceState);
		mBaseController.onCreate(this);
	}
	
	@Override
	protected void onDestroy() {
		super.onDestroy();
		mBaseController.onDestroy(this);
	}
	
	@Override
	protected void onStart() {
		super.onStart();
		mBaseController.mIsTop = true;
	}
	
	@Override
	protected void onStop() {
		super.onStop();
		mBaseController.mIsTop = false;
	}
	
	@Override
	protected void onResume() {
		super.onResume();
		mBaseController.onResume(this);
	}
	
	@Override
	protected void onPause() {
		super.onPause();
		mBaseController.onPause(this);
	}

	public void onEventMainThread(Message message){
		mBaseController.onEventMainThread(this, message);
	}
}
