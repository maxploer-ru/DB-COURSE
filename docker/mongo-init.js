const appDb = process.env.MONGO_APP_DB || "zvideo";
const appUser = process.env.MONGO_APP_USER || "zvideo_app";
const appPassword = process.env.MONGO_APP_PASSWORD || "zvideo_app_password";

const dbRef = db.getSiblingDB(appDb);

const existingUser = dbRef.getUser(appUser);
if (!existingUser) {
  dbRef.createUser({
    user: appUser,
    pwd: appPassword,
    roles: [{ role: "readWrite", db: appDb }],
  });
}

// TODO: refactor