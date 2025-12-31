import Movie from "../movie/Movie";

function Movies({ movies, updateMovieReview, message }) {
  return (
    <div className="container mt-4">
      <div className="row">
        {movies && movies.length > 0 ? (
          movies.map((m) => (
            <Movie
              key={m._id}
              movie={m}
              updateMovieReview={updateMovieReview}
            />
          ))
        ) : (
          <h2 className="">{message}</h2>
        )}
      </div>
    </div>
  );
}

export default Movies;
