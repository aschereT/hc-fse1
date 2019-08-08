import React, { Component } from "react";
import Col from 'react-bootstrap/Col';
import Form from 'react-bootstrap/Form';
import Button from 'react-bootstrap/Button';
import Modal from 'react-bootstrap/Modal'

class Mesform extends Component {

    constructor(props) {
        super(props);
        this.state = {
            username: '',
            message: '',
            notify: false,
            notHeader: '',
            notMessage: ''
        }
        this.handleChange = this.handleChange.bind(this);
        this.handleSubmit = this.handleSubmit.bind(this);
    }

    handleChange = e => this.setState({ [e.target.name]: e.target.value })

    handleSubmit(event) {
        event.preventDefault();
        let tt = this;
        fetch('http://localhost:8000/hcfse/post', {
            method: 'POST',
            headers: {
                Accept: 'application/json',
                'Content-Type': 'application/json',
            },
            body: JSON.stringify({
                username: this.state.username,
                content: this.state.message
            }),
        }).then(response => response.json()).then(function(response) {
            switch (response.Status) {
                case 406:
                    //too offensive
                    tt.setState((state, _) => {
                        return {notify: true, notHeader: "Failed to post message", notMessage: "Your message has been deemed too offensive. Please be civil and try again!"};
                    })
                    break;
                case 403:
                    //banned
                    tt.setState((state, _) => {
                        return {notify: true, notHeader: "Banned!", notMessage: "The username " + state.username + " has been banned for posting too many offensive messages!"};
                    })
                    break;
                default:
                    //success!
                    tt.setState((state, _) => {
                        return {notify: true, notHeader: "Success!", notMessage: "Your message has been posted!", message: '', username: state.username};
                    });
                    break;
            }})
    }

    render() {
        if(this.state.notify) {
            return (
                <Modal show={this.state.notify} onHide={() => this.setState((state, _) => {
                    return {notify: false};
                })}>
                  <Modal.Header closeButton>
                    <Modal.Title>{this.state.notHeader}</Modal.Title>
                  </Modal.Header>
                  <Modal.Body>{this.state.notMessage}</Modal.Body>
                  <Modal.Footer>
                    <Button variant="primary" onClick={() => this.setState((state, _) => {
                    return {notify: false};
                })}>
                      Close
                    </Button>
                  </Modal.Footer>
                </Modal>
            );
        }
        return (
            <Form onSubmit={this.handleSubmit}>
                <br></br>
                <Form.Row>
                    <Col>
                        <Form.Group controlId="formUsername">
                            <Form.Control name="username" type="text" placeholder="Enter your username" value={this.state.username} onChange={this.handleChange} />
                        </Form.Group>
                    </Col>
                    <Col>
                        <Form.Group controlId="formMessage">
                            <Form.Control name="message" type="text" placeholder="Enter your message" value={this.state.message} onChange={this.handleChange} />
                        </Form.Group>
                    </Col>
                    <Button variant="primary" type="submit" value="submit" lg={1}>
                        Submit
                </Button>
                </Form.Row>
            </Form>)
    }
}
export { Mesform }