import React, { Component } from "react";
import ListGroup from 'react-bootstrap/ListGroup';
import { resolve } from "dns";


class MesList extends Component {
    constructor() {
        super();
        this.state = {
            messages: {}
        }
        this.timer = setInterval(() => this.getBulk(), 1000);
    }

    getBulk() {
        fetch('http://localhost:8000/hcfse/bulk')
            .then(response => response.json(), _ => new Promise(() => resolve({BulkMes: {}})))
            .then(data => this.setState({messages: data.BulkMes})
            )
    }

    componentDidMount() {
        this.getBulk()
    }

    renderMes(message, index) {
        return (
            <ListGroup.Item key={index}>
                <div className="media">
                    <div className="media-body">
                        <h5 className="mt-0">Username: {message.username}</h5>
                        {message.content}
                    </div>
                </div>
            </ListGroup.Item>
        )
    }

    render() {
        // console.log(this.state)
        if (this.state.messages === undefined) {
            return (
                <ListGroup>
                </ListGroup>
            )
        }
        return (
            <ListGroup>
                {Object.values(this.state.messages).map(this.renderMes)}
            </ListGroup>
        );
    }
}
export { MesList }