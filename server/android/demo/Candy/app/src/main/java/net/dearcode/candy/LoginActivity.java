package net.dearcode.candy;

import android.content.Intent;
import android.os.Bundle;
import android.support.v7.app.AppCompatActivity;
import android.view.View;
import android.view.View.OnClickListener;
import android.widget.AutoCompleteTextView;
import android.widget.Button;
import android.widget.EditText;


/**
 * A login screen that offers login via email/password.
 */
public class LoginActivity extends AppCompatActivity {

    // UI references.
    private AutoCompleteTextView mEmailView;
    private EditText mPasswordView;
    private LoginActivity self;

    @Override
    protected void onCreate(Bundle savedInstanceState) {
        super.onCreate(savedInstanceState);
        self = this;
        setContentView(R.layout.activity_login);
        // Set up the login form.
        mEmailView = (AutoCompleteTextView) findViewById(R.id.email);

        mPasswordView = (EditText) findViewById(R.id.password);

        Button mEmailSignInButton = (Button) findViewById(R.id.email_sign_in_button);
        mEmailSignInButton.setOnClickListener(new OnClickListener() {
            @Override
            public void onClick(View view) {
                Intent ri = new Intent(LoginActivity.this, MainActivity.class);
                Bundle bundle = new Bundle();
                String email = mEmailView.getText().toString();
                String password = mPasswordView.getText().toString();
                bundle.putString("user", email);
                bundle.putString("pass", password);
                ri.putExtras(bundle);
                self.setResult(RESULT_OK, ri); //这理有2个参数(int resultCode, Intent intent)
                finish();
            }
        });
    }
}

