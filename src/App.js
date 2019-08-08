import React, { Component } from "react";
import Container from 'react-bootstrap/Container';
import Row from 'react-bootstrap/Row';
import { MesList } from './Meslist';
import { Mesform } from "./Mesform";
import { ShowHide } from "./ShowHide";

// Importing the Bootstrap CSS
// import "bootstrap/dist/css/bootstrap.min.css";
// Darkly theme from https://bootswatch.com/solar/
import "./bootstrap.min.css"
import "./App.css";


class App extends Component {
  constructor(props) {
    super(props);
    this.state = {
      showForm: false
    }
    this.showFlip = this.showFlip.bind(this)
  }

  showFlip() {
    console.log("Flip called, value is now " + this.state.showForm)
    this.setState({showForm: !this.state.showForm})
  }

  render() {
    return (
      <>
        <Container>
          <Row>
            <MesList></MesList>
          </Row>
        </Container>
        <ShowHide flipit={this.showFlip}></ShowHide>
        <div className="form">
          <Mesform showit={this.state.showForm}></Mesform>
        </div>
      </>
    );
  }
}

export default App;
