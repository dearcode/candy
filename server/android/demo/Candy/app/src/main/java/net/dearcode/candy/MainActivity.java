package net.dearcode.candy;

import android.content.ComponentName;
import android.content.Context;
import android.content.Intent;
import android.content.ServiceConnection;
import android.database.Cursor;
import android.database.sqlite.SQLiteDatabase;
import android.os.Bundle;
import android.os.IBinder;
import android.support.design.widget.FloatingActionButton;
import android.support.design.widget.NavigationView;
import android.support.design.widget.Snackbar;
import android.support.v4.view.GravityCompat;
import android.support.v4.widget.DrawerLayout;
import android.support.v7.app.ActionBarDrawerToggle;
import android.support.v7.app.AppCompatActivity;
import android.support.v7.widget.Toolbar;
import android.util.Log;
import android.view.Menu;
import android.view.MenuItem;
import android.view.View;
import android.widget.TextView;
import android.widget.Toast;

import net.dearcode.candy.model.ServiceResponse;
import net.dearcode.candy.service.Message;


public class MainActivity extends AppCompatActivity implements NavigationView.OnNavigationItemSelectedListener {
    private CounterServiceConnection conn = null;
    private static final String TAG = "CandyMessage";
    private long id;
    private String user;
    private String pass;
    private TextView tvUserName;
    private TextView tvUserID;
    private SQLiteDatabase db;

    @Override
    protected void onCreate(Bundle savedInstanceState) {
        super.onCreate(savedInstanceState);
        db = openOrCreateDatabase("candy.db", Context.MODE_PRIVATE, null);
        db.execSQL("CREATE TABLE IF NOT EXISTS user (id INTEGER, user TEXT, pass TEXT)");

        Cursor c = db.rawQuery("SELECT id, user,pass FROM user limit 1", null);
        if (c.moveToNext()) {
            id = c.getLong(0);
            user = c.getString(1);
            pass = c.getString(2);
        }
        c.close();


        conn = new CounterServiceConnection();
        Intent i = new Intent(MainActivity.this, Message.class);
        Log.e(TAG, "bind:" + bindService(i, conn, Context.BIND_AUTO_CREATE));
        setContentView(R.layout.activity_main);
        Toolbar toolbar = (Toolbar) findViewById(R.id.toolbar);
        setSupportActionBar(toolbar);

        FloatingActionButton fab = (FloatingActionButton) findViewById(R.id.fab);
        fab.setOnClickListener(new View.OnClickListener() {
            @Override
            public void onClick(View view) {
                Snackbar.make(view, "Replace with your own action", Snackbar.LENGTH_LONG)
                        .setAction("Action", null).show();
            }
        });

        DrawerLayout drawer = (DrawerLayout) findViewById(R.id.drawer_layout);
        ActionBarDrawerToggle toggle = new ActionBarDrawerToggle(
                this, drawer, toolbar, R.string.navigation_drawer_open, R.string.navigation_drawer_close);
        drawer.setDrawerListener(toggle);
        toggle.syncState();

        NavigationView navigationView = (NavigationView) findViewById(R.id.nav_view);
        navigationView.setNavigationItemSelectedListener(this);
        tvUserName = (TextView) navigationView.findViewById(R.id.tvUserName);
        tvUserID = (TextView) navigationView.findViewById(R.id.tvUserID);
        //如果没有保存的账号，就让他登录
        if (user == null || pass == null || user.isEmpty() || pass.isEmpty()) {
            Intent bintent = new Intent(MainActivity.this, LoginActivity.class);
            startActivityForResult(bintent, waitLogin);
        }else {
            tvUserName.setText(user);
            tvUserID.setText("ID：" + id);

        }
    }

    @Override
    public void onBackPressed() {
        DrawerLayout drawer = (DrawerLayout) findViewById(R.id.drawer_layout);
        if (drawer.isDrawerOpen(GravityCompat.START)) {
            drawer.closeDrawer(GravityCompat.START);
        } else {
            super.onBackPressed();
        }
    }

    @Override
    public boolean onCreateOptionsMenu(Menu menu) {
        // Inflate the menu; this adds items to the action bar if it is present.
        getMenuInflater().inflate(R.menu.main, menu);
        return true;
    }

    private static final int waitLogin = 898;
    private static final int waitRegister = 1;

    private long login(String user, String pass) throws Exception {
        ServiceResponse sr;
        try {
            sr = remoteService.login(user, pass);
            if (sr.hasError) {
                throw new Exception(sr.getError());
            }
        } catch (Exception e) {
            throw new Exception(e.getMessage());
        }
        return sr.getId();
    }

    private long register(String user, String pass) throws Exception {
        ServiceResponse sr;
        try {
            sr = remoteService.register(user, pass);
            if (sr.hasError) {
                throw new Exception(sr.getError());
            }
        } catch (Exception e) {
            throw new Exception(e.getMessage());
        }
        return sr.getId();
    }

    protected void onActivityResult(int requestCode, int resultCode, Intent data) {
        if (resultCode != RESULT_OK)
            return;
        Bundle b = data.getExtras();
        long id;

        switch (requestCode) {
            case waitLogin:
                try {
                    id = this.login(b.getString("user"), b.getString("pass"));
                    showInfo("user login success id:" + id);
                    db.execSQL("insert into user (id, user,pass) values (?,?,?)", new Object[]{id, b.getString("user"), b.getString("pass")});
                } catch (Exception e) {
                    if (e.getMessage().contains("not found")) {
                        Intent bintent = new Intent(MainActivity.this, RegisterActivity.class);
                        startActivityForResult(bintent, waitRegister);
                    }
                    showInfo(e.getMessage());
                    Log.e(TAG, e.getMessage());
                }
                break;
            case waitRegister:
                try {
                    id = this.register(b.getString("user"), b.getString("pass"));
                    showInfo("user register success id:" + id);
                    tvUserName.setText(b.getString("user"));
                    tvUserID.setText("ID：" + id);

                } catch (Exception e) {
                    showInfo("user register error:" + e.getMessage());
                }
                break;
            default:
                break;
        }
    }

    @Override
    public boolean onOptionsItemSelected(MenuItem item) {
        // Handle action bar item clicks here. The action bar will
        // automatically handle clicks on the Home/Up button, so long
        // as you specify a parent activity in AndroidManifest.xml.
        int id = item.getItemId();

        //noinspection SimplifiableIfStatement
        if (id == R.id.action_settings) {
            Intent bintent = new Intent(MainActivity.this, LoginActivity.class);
            startActivityForResult(bintent, 0);
            return true;
        }

        return super.onOptionsItemSelected(item);
    }

    @SuppressWarnings("StatementWithEmptyBody")
    @Override
    public boolean onNavigationItemSelected(MenuItem item) {
        // Handle navigation view item clicks here.
        int id = item.getItemId();

        if (id == R.id.nav_camera) {
            // Handle the camera action
        } else if (id == R.id.nav_gallery) {

        } else if (id == R.id.nav_slideshow) {

        } else if (id == R.id.nav_manage) {

        } else if (id == R.id.nav_share) {

        } else if (id == R.id.nav_send) {

        }

        DrawerLayout drawer = (DrawerLayout) findViewById(R.id.drawer_layout);
        drawer.closeDrawer(GravityCompat.START);
        return true;
    }

    private void showInfo(String msg) {
        Toast.makeText(this, msg, Toast.LENGTH_SHORT).show();
    }

    private CandyMessage remoteService = null;

    private class CounterServiceConnection implements ServiceConnection {
        @Override
        public void onServiceConnected(ComponentName name, IBinder service) {
            remoteService = CandyMessage.Stub.asInterface(service);
            showInfo("onServiceConnected");
        }

        @Override
        public void onServiceDisconnected(ComponentName name) {
            remoteService = null;
            showInfo("onServiceDisconnected");
        }
    }
}
