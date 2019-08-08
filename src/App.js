import React from "react";
import Container from 'react-bootstrap/Container';
import Row from 'react-bootstrap/Row';
import { MesList } from './Meslist';
import { Mesform } from "./Mesform";

// Importing the Bootstrap CSS
// import "bootstrap/dist/css/bootstrap.min.css";
// Darkly theme from https://bootswatch.com/solar/
import "./bootstrap.min.css"
import "./App.css";


const App = () => (
  <>
  <Container>
      <Row>
        <MesList></MesList>
      </Row>
  </Container>
  <Mesform></Mesform>
  </>
);

export default App;
