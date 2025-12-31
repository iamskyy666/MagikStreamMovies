import Home from "./components/home/Home";
import Header from "./components/header/Header";
import { Route, Routes, useNavigate } from "react-router-dom";
import Register from "./components/register/Register";
import Login from "./components/login/Login";
import Layout from "./components/Layout";
import RequiredAuth from "./components/RequiredAuth";
import Recommended from "./components/recommneded/Recommended";
import Review from "./components/review/Review";
import useAuth from "./hooks/useAuth";
import axiosClient from "./api/axiosConfig";
import StreamMovie from "./components/stream/StreamMovie";

function App() {
  const navigate = useNavigate();
  const { auth, setAuth } = useAuth();

  const updateMovieReview = (imdb_id) => {
    navigate(`/review/${imdb_id}`);
  };

  const handleLogout = async () => {
    try {
      const response = await axiosClient.post("/logout", {
        user_id: auth.user_id,
      });
      console.log(response.data);
      setAuth(null);
      // localStorage.removeItem('user');
      console.log("User logged out");
    } catch (error) {
      console.error("Error logging out:", error);
    }
  };

  return (
    <>
      <Header handleLogout={handleLogout}/>
      <Routes path="/" element={<Layout />}>
        {/* Unprotected Routes ➡️*/}
        <Route
          path="/"
          element={<Home updateMovieReview={updateMovieReview} />}
        />
        <Route path="/register" element={<Register />} />
        <Route path="/login" element={<Login />} />
        {/* Protected Routes 🛡️*/}
        <Route element={<RequiredAuth />}>
          <Route path="/recommended" element={<Recommended />} />
          <Route path="/review/:imdb_id" element={<Review />} />
          <Route path="/stream/:yt_id" element={<StreamMovie />} />
        </Route>
      </Routes>
    </>
  );
}

export default App;

//14:

// mongodb+srv://<db_username>:<db_password>@magik-stream-movies.cwkqncn.mongodb.net/?appName=magik-stream-movies