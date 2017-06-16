import {Component} from '@angular/core'
import {Router, ActivatedRoute, Params} from '@angular/router'
import {Observable} from 'rxjs/Rx'

import {JoDirectService, JoDirectMessages, JoDirectCode} from '../../services/jodirect.service'
import {Utils} from '../../utils/utils'

@Component({
    selector: `jodirect-messages-page`,
    template: `
        <div>
            <div id="parent" [style.display]="messagesDisplay === 'none' ? 'block': 'none'">
                <span style="color:red;font-style: italic;font-size: 17px">{{error}}</span>
            </div>

            <pageparent-view [style.display]="messagesDisplay">
                <span style="font-size: 18px;margin:16px"><b>Token {{token}} will expire in {{time}}</b></span><br>
                <span [style.display]="messages.length == 0 ? 'block' : 'none'"
                      style="margin-left:16px">No messages</span>
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
    time: string
    error: string
    messages = []

    timeleft: number
    timer: any

    messagesDisplay = 'none'

    constructor(private router: Router,
                private route: ActivatedRoute,
                private joDirectService: JoDirectService) {
    }

    ngOnInit() {
        this.route.queryParams.forEach((params: Params) => {
            if (params.hasOwnProperty('token') && params.hasOwnProperty('password')) {
                this.joDirectService.login(params.token, params.password).forEach((messages: JoDirectMessages) => {
                    if (messages.success()) {
                        this.messagesDisplay = 'block'
                        this.token = messages.getToken()
                        this.timeleft = Math.ceil(messages.getTimeleft())
                        this.messages = messages.getMessages()

                        this.timer = Observable.timer(0, 1000).subscribe(() => {
                            this.timeleft--
                            if (this.timeleft >= 0) {
                                this.updateTime()
                            }
                        })
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

    updateTime() {
        this.time = Utils.formatSeconds(this.timeleft)
    }

    ngOnDestroy() {
        if (this.timer != null) {
            this.timer.unsubscribe()
        }
    }

}
