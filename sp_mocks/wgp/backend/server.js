const express = require("express");
const cors = require("cors");

const jwt = require("jsonwebtoken");
const bcrypt = require("bcrypt");
const app = express();
const {poems, users} = require("./mock-database.js") 
const SECRET_KEY = "my_very_secret_key";
// TODO: For future, use a layer8 npm published package for initialising the client and variables
const popsicle = require("popsicle");
const ClientOAuth2 = require("client-oauth2");
require("dotenv").config();
const port = process.env.PORT;
const FRONTEND_URL = process.env.FRONTEND_URL;
const LAYER8_URL = process.env.LAYER8_URL;
// const port = 8000;
// const FRONTEND_URL = "http://localhost:5173"
// const LAYER8_URL = "http://localhost:5001"
const LAYER8_CALLBACK_URL = `${FRONTEND_URL}/oauth2/callback`;
const LAYER8_RESOURCE_URL = `${LAYER8_URL}/api/user`;

const layer8Auth = new ClientOAuth2({
  clientId: "f0fe2f3f-cabe-4d44-a8d4-ed7154627867",
  clientSecret: "b85bc41f06a0f86b912a51a9688a3c78fb464e9b7ac692eff145f20e4bcae3e8",
  accessTokenUri: `${LAYER8_URL}/api/oauth`,
  authorizationUri: `${LAYER8_URL}/authorize`,
  redirectUri: LAYER8_CALLBACK_URL,
  scopes: ["read:user"],
});

const layer8_middleware = require("layer8_middleware")

const upload = layer8_middleware.multipart({ dest: "pictures/dynamic" });

app.get("/healthcheck", (req, res) => {
  console.log("Enpoint for testing");
  console.log("req.body: ", req.body);
  res.send("Bro, ur poems coming soon. Relax a little.");
});

//const Layer8 = require("./dist/loadWASM.js");
//app.use(Layer8);

app.use(layer8_middleware.tunnel);

app.use(cors());
app.use('/media', layer8_middleware.static('pictures'));
app.use('/test', (req, res) => {
  res.status(200).json({ message: 'Test endpoint' });
});

app.post("/", (req, res) => {
  console.log("Enpoint for testing");
  // console.log("headers: ", req.headers);
  // console.log("req.body: ", req.body);
  res.setHeader("x-header-test", "1234");
  res.send("Server has registered a POST.");
});

let counter = 0;
app.get("/nextpoem", (req, res) => {
  counter++;
  let marker = counter % 3;
  console.log("Served: ", poems[marker].title);
  res.status(200).json(poems[marker]);
});

app.post("/api/register", async (req, res) => {
  const { password, email, profile_image } = req.body;

  try {
    const hashedPassword = await bcrypt.hash(password, 10);
    users.push({ email, password: hashedPassword, profile_image });
    res.status(200).send("User registered successfully!");
  } catch (err) {
    console.log("err: ", err);
    res.status(500).send({ error: "Something went wrong!" });
  }
});

app.post("/api/login", async (req, res) => {
  //console.log("users: ", users);
  const { email, password } = req.body;
  const user = users.find((u) => u.email === email);
  if (user && (await bcrypt.compare(password, user.password))) {
    const token = jwt.sign({ email }, SECRET_KEY);
    res.status(200).json({ user, token });
  } else {
    res.status(401).json({ error: "Invalid credentials!" });
  }
});

// Layer8 Components start here
app.get("/api/login/layer8/auth", async (req, res) => {
  console.log("layer8Auth.code.getUri(): ", layer8Auth.code.getUri());
  res.status(200).json({ authURL: layer8Auth.code.getUri() });
});

app.post("/api/login/layer8/auth", async (req, res) => {
  const { callback_url } = req.body;
  const user = await layer8Auth.code
    .getToken(callback_url)
    .then(async (user) => {
      return await popsicle
        .request(
          user.sign({
            method: "GET",
            url: LAYER8_RESOURCE_URL,
          })
        )
        .then((res) => {
          //console.log("response: ", res);
          return JSON.parse(res.body);
        })
        .catch((err) => {
          console.log("Error from pkg Popsicle: ", err);
        });
    })
    .catch((err) => {
      console.log("err: ", err);
    });

  const isEmailVerified = user.is_email_verified.value;
  let displayName = "";
  let countryName = "";

  // Metadata for Layer8
  let Sec_Ch_Ua_Platform = "";
  let Sec_Fetch_Site = "";
  let Referer = "";
  let Sec_Ch_Ua = "";
  let User_Agent = "";

  if (user.display_name) {
    displayName = user.display_name.value;
  }

  if (user.country_name) {
    countryName = user.country_name.value;
  }

  // Metadata for Layer8
  if (user.hm_sec_ch_ua_platform) {
    Sec_Ch_Ua_Platform = user.hm_sec_ch_ua_platform;
    Sec_Fetch_Site = user.hm_sec_fetch_site;
    Referer = user.hm_referer;
    Sec_Ch_Ua = user.hm_sec_ch_ua;
    User_Agent = user.hm_user_agent;
    console.log("Sec_Ch_Ua_Platform: ", Sec_Ch_Ua_Platform);
    console.log("Sec_Fetch_Site: ", Sec_Fetch_Site);
    console.log("Referer: ", Referer);
    console.log("Sec_Ch_Ua: ", Sec_Ch_Ua);
    console.log("User_Agent: ", User_Agent);
  }

  const token = jwt.sign(
    { isEmailVerified, displayName, countryName },
    SECRET_KEY
  );
  res.status(200).json({ token });
});

app.post("/api/profile/upload", upload.single('file'), (req, res) => {
  const uploadedFile = req.file;

  if (!uploadedFile) {
    return res.status(400).json({ error: 'No file uploaded' });
  }

  res.status(200).json({ 
    message: "File uploaded successfully!",
    url: `${req.protocol}://${req.get('host')}/media/dynamic/${req.file?.name}`
  });
});


app.listen(port, () => {
  console.log(
    `\nA mock Service Provider backend is now listening on port ${port}.`
  );
});
