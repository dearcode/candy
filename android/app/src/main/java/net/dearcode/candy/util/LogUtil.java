/*
 * Copyright (C)  Tony Green, Litepal Framework Open Source Project
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *      http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package net.dearcode.candy.util;

import android.content.Context;
import android.util.Log;

import com.forlong401.log.transaction.log.manager.LogManager;
import com.forlong401.log.transaction.utils.LogUtils;
/**
 * Created by 水寒 on 2016/9/17.
 * 日志打印工具类
 */

public final class LogUtil {
	
	private static final String TAG_NAME = "XYS";
	
	public static final int DEBUG = 2;
	
	public static final int ERROR = 5;
	
	public static int level = DEBUG;
	
	public static void d(String tagName, String message) {
		if (level <= DEBUG) {
			if(message == null) return;
			Log.d(tagName, message);
		}
	}
	
	public static void e(String tagName, Exception e){
		if (level <= ERROR) {
			Log.e(tagName, e.getMessage(), e);
		}
	}
	
	public static void e(String message){
		e(TAG_NAME, message);
	}
	
	public static void e(String tagName, String message){
		if(level <= ERROR){
			if(message == null) return;
			Log.e(tagName, message);
		}
	}
	
	public static void d(String message){
		d(TAG_NAME, message);
	}
	
	public static void e(Exception e){
		e(TAG_NAME, e);
	}
	
	/**
	 * 日志收集到 SD card 目录下的包名中点替换为下划线的文件夹下的 + log/crash + 日志文件
	 * 详细请看： https://github.com/licong/log
	 * @param context
	 * @param message
	 */
	public static void logFile(Context context, String message){
		logFile(context, TAG_NAME, message);
	}
	
	public static void logFile(Context context, String tag, String message){
		LogManager.getManager(context.getApplicationContext())
		.log(tag, message, LogUtils.LOG_TYPE_2_FILE_AND_LOGCAT);
	}
}