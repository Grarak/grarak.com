import { Component, HostListener } from '@angular/core'

import { Utils } from '../utils/utils'

@Component({
    selector: `aboutme-page`,
    template: `
        <pageparent-view>
            <card-view>
                <card-title><span style="font-size:x-large;">Welcome</span></card-title>
                <card-content>
                    <div style="font-size:larger;color:rgba(0,0,0,.7)">
                        Hello! I'm Willi Ye aka Grarak, a 19 year old student at Vienna University of Technology.<br><br>

                        In my free time I contribute to OSS projects such as
                        <a target="_blank" href="https://github.com/CyanogenMod">CyanogenMod</a>/<a target="_blank" href="https://github.com/LineageOS">LineageOS</a>.
                        All my personal projects are uploaded on github as well. You can find a link to my profile down below.<br><br>

                        I've been writing software since I was 15 years old. Everything started when I decided to work on Android ROMs. After graduating from High School,
                        I decided to go to university to get a BSc in Software Engineering.
                    </div>

                    <div id="links">
                        <a class="linkicon" target="_blank" href="{{link.link}}" *ngFor="let link of links" style="margin-left:5px;margin-right:5px">
                            <img [src]="link.icon" width="25">
                        </a>
                    </div>
                </card-content>
            </card-view>
        </pageparent-view>
    `,
    styles: [
        `
            #parent {
                margin: 30px;
            }

            h1 {
                color: rgba(0, 0, 0, .70);
            }

            #links {
                padding-top: 20px;
                display: flex;
                justify-content: center;
            }

            .linkicon {
                float: left;
            }
        `
    ]
})
export class AboutMeComponent {

    links: any[] = [
        { link: "https://twitter.com/Grarak", icon: Utils.getAsset('ic_twitter.png') },
        { link: "https://plus.google.com/u/0/101665005278935743165", icon: Utils.getAsset('ic_google_plus.png') },
        { link: "https://github.com/Grarak", icon: Utils.getAsset('ic_github.png') }
    ]

}
