import {Component} from '@angular/core'

import {JoDirectService, JoDirectResponse, JoDirectCode} from '../../services/jodirect.service'

@Component({
    selector: `jodirect-send-page`,
    template: `
        <div id="parent">
            <span [style.display]="itemDisplay">You can only send one message to a token, so make sure you don't make a mistake.</span><br>
            <md-input-container [style.display]="itemDisplay">
                <input mdInput placeholder="Token" [(ngModel)]="token">
            </md-input-container>
            <md-input-container [style.display]="itemDisplay">
                <textarea mdInput cols="55" rows="15" placeholder="Message" [(ngModel)]="message"></textarea>
            </md-input-container>
            <span [style.display]="errorDisplay" style="color:red;font-style: italic;font-size: 14px">{{error}}</span>
            <br>
            <button md-raised-button color="accent" class="button" (click)="onSend()" [style.display]="itemDisplay">
                Send
            </button>
            <span [style.display]="successDisplay">Successfully sent message!</span>
        </div>
    `,
    styles: [`
        #parent {
            position: absolute;
            transform: translateX(-50%) translateY(-50%);
            left: 50%;
            top: 50%;
            text-align: center;
        }

        .button {
            margin-top: 10px;
            margin-bottom: 10px;
        }
    `]
})
export class JoDirectSendComponent {

    token = ''
    message = ''
    error = ''

    errorDisplay = 'none'
    itemDisplay = 'block'
    successDisplay = 'none'

    constructor(private joDirectService: JoDirectService) {
    }

    onSend() {
        this.error = ''
        this.errorDisplay = 'none'
        if (this.token === '') {
            this.error = 'Token is empty!'
        } else if (this.message === '') {
            this.error = 'Message is empty!'
        }

        if (this.error !== '') {
            this.errorDisplay = 'block'
        } else {
            this.joDirectService.send(this.token, this.message).forEach((response: JoDirectResponse) => {
                if (response.status() === JoDirectCode.NO_ERROR) {
                    this.itemDisplay = 'none'
                    this.errorDisplay = 'none'
                    this.successDisplay = 'block'
                } else if (response.status() === JoDirectCode.INVALID_TOKEN) {
                    this.error = this.token + ' does not exists!'
                } else if (response.status() === JoDirectCode.MESSAGE_ALREADY_SENT) {
                    this.error = 'You already sent a message to this token!'
                }

                if (this.error !== '') {
                    this.errorDisplay = 'block'
                }
            })
        }
    }

}
