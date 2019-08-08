import React, { Component } from "react";
import { Container, Button } from 'react-floating-action-button'


class ShowHide extends Component {
    constructor(props) {
        super(props);
    }

    render() {
        return (
            <Container>
                <Button
                    tooltip="Post a message"
                    icon="fas fa-plus"
                    rotate={false}
                    onClick={this.props.flipit} />
            </Container>
        )
    }

}

export { ShowHide }