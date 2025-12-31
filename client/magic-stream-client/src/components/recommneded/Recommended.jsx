import { useEffect, useState } from "react";
import useAxiosPrivate from "../../hooks/useAxiosPrivate";
import Movies from "../movies/Movies";

function Recommended() {
  const [movies, setMovies] = useState([]);
  const [loading, setLoading] = useState(false);
  const [message, setMessage] = useState();
  const axiosPrivate = useAxiosPrivate();

  useEffect(() => {
    async function fetchRecommendedMovies() {
      setLoading(true);
      setMessage("");
      try {
        const resp = await axiosPrivate.get("/recommended-movies");
        setMovies(resp.data);
      } catch (err) {
        console.log(`ERROR fetching recommended movies: ${err}`);
      } finally {
        setLoading(false);
      }
    }
    fetchRecommendedMovies();
  }, []);
  return (
    <>
      {loading ? (
        <h2>loading... ⌛</h2>
      ) : (
        <Movies movies={movies} message={message} />
      )}
    </>
  );
}

export default Recommended;
