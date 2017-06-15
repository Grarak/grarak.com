import {Component} from '@angular/core'
import {Router, ActivatedRoute, Params} from '@angular/router'

import {JoDirectService, JoDirectMessages, JoDirectCode} from '../../services/jodirect.service'

@Component({
    selector: `jodirect-messages-page`,
    template: `
        <div>
            <div id="parent" [style.display]="messagesDisplay === 'none' ? 'block': 'none'">
                <span style="color:red;font-style: italic;font-size: 17px">{{error}}</span>
            </div>

            <pageparent-view [style.display]="messagesDisplay">
                <span style="font-size: 20px"><b>Token: {{token}} expire in {{timeleft}}.</b></span><br>
                <span [style.display]="messages.length == 0 ? 'block' : 'none'">No messages</span>
                <div style="margin-top:10px;margin-bottom:10px" *ngFor="let m of messages; let i = index">
                    <card-view>
                        <card-content>
                            <div style="word-wrap:break-word">{{i + 1}}. {{m}}</div>
                        </card-content>
                    </card-view>
                </div>
            </pageparent-view>

        </div>
    `,
    styles: [
            `
            #parent {
                position: absolute;
                transform: translateX(-50%) translateY(-50%);
                left: 50%;
                top: 50%;
                text-align: center;
            }
        `
    ]
})
export class JoDirectMessagesComponent {

    token: string
    timeleft: string
    error: string
    messages = []

    messagesDisplay = 'none'

    constructor(private router: Router,
                private route: ActivatedRoute,
                private joDirectService: JoDirectService) {
    }

    ngOnInit() {
        this.route.queryParams.forEach((params: Params) => {
            if (params.hasOwnProperty('token') && params.hasOwnProperty('password')) {
                this.joDirectService.login(params.token, params.password).forEach((messages: JoDirectMessages) => {
                    console.log(messages)
                    if (messages.success()) {
                        this.messagesDisplay = 'block'
                        this.token = messages.getToken()

                        const timeleft: number = messages.getTimeleft()
                        const mins = Math.floor(timeleft / 60)
                        const seconds = Math.ceil(timeleft % 60)

                        let minsText = mins.toString()
                        let secondsText = seconds.toString()
                        if (mins < 10) {
                            minsText = '0' + minsText
                        }
                        if (seconds < 10) {
                            secondsText = '0' + secondsText
                        }
                        this.timeleft = '00:' + minsText + ':' + secondsText

                        this.messages = messages.getMessages()
                    } else if (messages.status() === JoDirectCode.INVALID_TOKEN
                        || messages.status() === JoDirectCode.WRONG_PASSWORD) {
                        this.error = 'Token invalid or password wrong!'
                    } else if (messages.status() === JoDirectCode.TOKEN_EXPIRED) {
                        this.error = 'Token expired!'
                    }
                })
            } else {
                this.router.navigate(['404'])
            }
        })
    }

}
