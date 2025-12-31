import { Button, Container, Nav, Navbar } from "react-bootstrap";
import { NavLink, useNavigate } from "react-router-dom";
import useAuth from "../../hooks/useAuth";
import logo from '../../assets/logo.png'

function Header({ handleLogout }) {
  const navigate = useNavigate();
  const { auth } = useAuth();
  //console.log("AUTH Obj:", auth);

  return (
    <Navbar bg="dark" variant="dark" expand="lg" className="shadow-sm">
      <Container>
        <Navbar.Brand>
          <img
            src={logo}
            alt="brand-logo"
            width="30"
            height="30"
            className="d-inline-block align-top me-2"
          />
          Magik Stream
        </Navbar.Brand>
        <Navbar.Toggle aria-controls="main-navbar-nav" />
        <Navbar.Collapse>
          <Nav className="me-auto">
            <Nav.Link as={NavLink} to="/">
              Home
            </Nav.Link>
            <Nav.Link as={NavLink} to="/recommended">
              Recommended
            </Nav.Link>
          </Nav>
          <Nav className="ms-auto align-items-center">
            {auth ? (
              <>
                <span className="me-3 text-light">
                  Hello, <strong>{auth.first_name}</strong>
                </span>
                <Button
                  variant="outline-light"
                  size="sm"
                  onClick={handleLogout}>
                  Logout
                </Button>
              </>
            ) : (
              <>
                <Button
                  variant="outline-info"
                  size="sm"
                  className="me-2"
                  onClick={() => navigate("/login")}>
                  Login
                </Button>
                <Button
                  variant="outline-info"
                  size="sm"
                  className="me-2"
                  onClick={() => navigate("/register")}>
                  Register
                </Button>
              </>
            )}
          </Nav>
        </Navbar.Collapse>
      </Container>
    </Navbar>
  );
}

export default Header;
