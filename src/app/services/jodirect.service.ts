import {Injectable} from '@angular/core'
import {Http, Response, URLSearchParams} from '@angular/http'

import {Observable} from 'rxjs/Observable'

export class JoDirectCode {
    public static NO_ERROR = 0
    public static INVALID_TOKEN = 1
    public static TOKEN_EXPIRED = 2
    public static MESSAGE_ALREADY_SENT = 3
    public static WRONG_PASSWORD = 4
}

export class JoDirectToken {

    private content: any

    constructor(private response: any) {
        this.content = response.content
    }

    getTimeleft(): number {
        return this.content.timeleft
    }

    getPassword(): string {
        return this.content.password
    }

    getToken(): string {
        return this.content.token
    }

    success(): boolean {
        return this.content.hasOwnProperty('token') && this.content.hasOwnProperty('password')
    }

}

export class JoDirectMessages {

    private content: any

    constructor(private response: any) {
        this.content = response.content
    }

    getMessages(): string[] {
        if (this.content.hasOwnProperty('messages')) {
            return this.content.messages
        }
        return []
    }

    getTimeleft(): number {
        return this.content.timeleft
    }

    getToken(): string {
        return this.content.token
    }

    status(): number {
        return this.content.status
    }

    success(): boolean {
        return this.content.hasOwnProperty('token') && this.response.success
    }
}

export class JoDirectResponse {

    private content: any

    constructor(private response: any) {
        this.content = response.content
    }

    status(): number {
        return this.content.status
    }

    success(): boolean {
        return this.content.hasOwnProperty('token') && this.response.success
    }

}

@Injectable()
export class JoDirectService {

    private genTokenLink = '/jodirect/api/v1/token/generate'
    private genMessageReceivedLink = '/jodirect/api/v1/message/received?'
    private sendMessageLink = '/jodirect/api/v1/message/send'

    constructor(private http: Http) {
    }

    genToken(): Observable<JoDirectToken> {
        return new Observable<JoDirectToken>((observer: any) => {
            this.http.get(this.genTokenLink).forEach((response: Response) => {
                if (response.status === 404) {
                    observer.next(null)
                } else {
                    observer.next(new JoDirectToken(response.json()))
                }
            })
        })
    }

    login(token: string, password: string): Observable<JoDirectMessages> {
        const url = new URLSearchParams(this.genMessageReceivedLink)
        url.set('token', token)
        url.set('password', password)
        return new Observable<JoDirectMessages>((observer: any) => {
            this.http.get(url.toString()).forEach((response: Response) => {
                if (response.status === 404) {
                    observer.next(null)
                } else {
                    observer.next(new JoDirectMessages(response.json()))
                }
            })
        })
    }

    send(token: string, message: string): Observable<JoDirectResponse> {
        const body = {token: token, content: message}
        return new Observable<JoDirectResponse>((observer: any) => {
            this.http.post(this.sendMessageLink, body).forEach((response: Response) => {
                if (response.status === 404) {
                    observer.next(null)
                } else {
                    observer.next(new JoDirectResponse(response.json()))
                }
            })
        })
    }

}
