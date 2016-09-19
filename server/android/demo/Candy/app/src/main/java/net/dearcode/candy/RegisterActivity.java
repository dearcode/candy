package net.dearcode.candy;

import android.content.Intent;
import android.os.Bundle;
import android.support.v7.app.AppCompatActivity;
import android.text.TextUtils;
import android.view.View;
import android.view.View.OnClickListener;
import android.widget.AutoCompleteTextView;
import android.widget.Button;
import android.widget.EditText;
import android.widget.Toast;


public class RegisterActivity extends AppCompatActivity {

    // UI references.
    private AutoCompleteTextView mEmail;
    private EditText mPassword;
    private EditText mPassword2;
    private RegisterActivity self;
    private String GetString(CharSequence cs) {
        if (TextUtils.isEmpty(cs)) {
            return "";
        }
        return cs.toString();
    }

    @Override
    protected void onCreate(Bundle savedInstanceState) {
        super.onCreate(savedInstanceState);
        self = this;
        setContentView(R.layout.activity_register);
        // Set up the login form.
        mEmail = (AutoCompleteTextView) findViewById(R.id.email);
        mPassword = (EditText) findViewById(R.id.password);
        mPassword2 = (EditText) findViewById(R.id.password2);

        Button mEmailSignInButton = (Button) findViewById(R.id.email_sign_in_button);
        mEmailSignInButton.setOnClickListener(new OnClickListener() {
            @Override
            public void onClick(View view) {
                Intent ri = new Intent(RegisterActivity.this, MainActivity.class);
                Bundle bundle = new Bundle();
                String email =  GetString( mEmail.getText());
                String password = GetString( mPassword.getText());
                String password2 = GetString(mPassword2.getText());
                if (TextUtils.isEmpty(email)) {
                    Toast.makeText(self, "用户名不能为空", Toast.LENGTH_SHORT).show();
                    return;
                }

                if (TextUtils.isEmpty(password) || TextUtils.isEmpty(password2)) {
                    Toast.makeText(self, "密码不能为空", Toast.LENGTH_SHORT).show();
                    return;
                }

                if (!TextUtils.equals(password, password2)) {
                    Toast.makeText(self, "两次密码不一致", Toast.LENGTH_SHORT).show();
                    return;
                }

                bundle.putString("user", email);
                bundle.putString("pass", password);
                ri.putExtras(bundle);
                self.setResult(RESULT_OK, ri); //这理有2个参数(int resultCode, Intent intent)
                finish();
            }
        });
    }
}

