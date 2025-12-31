import { useEffect, useState } from "react";
import axiosClient from "../../api/axiosConfig";
import Movies from "../movies/Movies";

function Home({ updateMovieReview }) {
  const [movies, setMovies] = useState([]);
  const [loading, setLoading] = useState(false);
  const [msg, setMsg] = useState(null);

  useEffect(() => {
    const fetchMovies = async () => {
      setLoading(true);
      setMsg("");
      try {
        const resp = await axiosClient.get("/movies");
        setMovies(resp.data);
        if (resp.data.length === 0) {
          setMsg("There are currently no movies available!");
        }
      } catch (error) {
        console.log(`⚠️ Error fetching movies: ${error}`);
        setMsg("⚠️ ERROR fetching movies!");
      } finally {
        setLoading(false);
      }
    };
    fetchMovies(); // ✅ THIS WAS MISSING
  }, []);
  return (
    <>
      {loading ? (
        <h2>Loading... ⌛</h2>
      ) : (
        <Movies
          movies={movies}
          message={msg}
          updateMovieReview={updateMovieReview}
        />
      )}
    </>
  );
}

export default Home;
