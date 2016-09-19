package net.dearcode.candy.model;

import android.os.Parcel;
import android.os.Parcelable;
import android.util.Log;

/**
 *  * Created by c-wind on 2016/9/19 14:31
 *  * mail：root@codecn.org
 *  
 */
public class ServiceResponse implements Parcelable {
    public boolean hasError;
    public long id;
    public String error;

    private static final String TAG = "CandyMessage";

    public static final Parcelable.Creator<ServiceResponse> CREATOR = new
            Parcelable.Creator<ServiceResponse>() {
                public ServiceResponse createFromParcel(Parcel in) {
                    return new ServiceResponse(in);
                }

                public ServiceResponse[] newArray(int size) {
                    return new ServiceResponse[size];
                }
            };

    public ServiceResponse() {
    }
    protected ServiceResponse(Parcel in) {
        this.error = in.readString();
        this.id = in.readLong();
        if (this.error != null && !this.error.isEmpty()) {
            this.hasError = true;
        }
        Log.e(TAG, "recover data id:"+this.id+" err:"+this.error+" ok:"+(this.error != null && !this.error.isEmpty()) );
    }

    public String getError() {
        Log.e(TAG, "getError:"+error);
        return error;
    }

    public void setError(String error) {
        this.hasError = true;
        this.error = error;
        Log.e(TAG, "setErr:"+error);
    }

    @Override
    public int describeContents() {
        return 0;
    }

    @Override
    public void writeToParcel(Parcel dest, int flags) {
        dest.writeString(this.error);
        dest.writeLong(this.id);
        Log.e(TAG, "write data");
    }

    public long getId() {
        return id;
    }

    public void setId(long id) {
        this.id = id;
    }
}
