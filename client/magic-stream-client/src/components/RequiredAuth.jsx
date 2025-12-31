import { Navigate, Outlet, useLocation } from "react-router-dom";
import useAuth from "../hooks/useAuth";
import Spinner from "./spinner/Spinner";

function RequiredAuth() {
  const { auth,loading } = useAuth();
  const location = useLocation();
  if (loading){
    return (<Spinner/>)
  }
  return auth ? (
    <Outlet />
  ) : (
    <Navigate to="/login" state={{ from: location }} replace />
  );
}

export default RequiredAuth;
