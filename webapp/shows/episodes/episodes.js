angular.module('shows.episodes',[
  'shows.episodes.create',
  'shows.episodes.edit',
  'tv2go.models.shows',
  'tv2go.models.episodes',
])
.config(function($stateProvider){
  $stateProvider
    .state('tv2go.shows.episodes', {
      url: '/:show',
      views: {
        'detail@tv2go': {
          templateUrl: 'shows/episodes/episodes.tmpl.html',
          controller: 'EpisodesListCtrl as episodesListCtrl',
        },
        'showdetail@tv2go.shows.episodes': {
          templateUrl: 'shows/showdetail.tmpl.html',
          controller: 'EpisodesListCtrl as episodesListCtrl',
        }
      },
      resolve: {
        show: function($stateParams, ShowsModel){
          return ShowsModel.getShowById($stateParams.show);
        },
        eps: function($stateParams, EpisodesModel) {
          return EpisodesModel.getEpisodes($stateParams.show);
        }
      },
    });
})
.controller('EpisodesListCtrl', function ($stateParams, show, eps, EpisodesModel, ShowsModel, ModalService) {
  var EpisodesListCtrl = this;
  
  EpisodesListCtrl.episodes = eps;
  EpisodesListCtrl.show = show;
  EpisodesListCtrl.statuses = ['WANTED', 'SKIPPED'];

  EpisodesListCtrl.updateStatus = function(ep) {
    EpisodesModel.updateEpisode(ep);
  };

  EpisodesListCtrl.updateShow = function(show) {
    console.log(show);
    show.$updateFromIndexer({id:show.id});
  };
  EpisodesListCtrl.updateShowFromDisk = function(show) {
    console.log(show);
    show.$updateFromDisk({id:show.id});
  };

  EpisodesListCtrl.saveShow = function(show) {
    show.$update({showid:show.id});
  };
  EpisodesListCtrl.searchEpisode = function(ep) {
    ModalService.showModal({
      templateUrl: 'shows/episodes/modal.tmpl.html',
      controller: "EpisodeSearchCtrl",
      controllerAs: "episodeSearchCtrl",
      inputs: {
        ep: ep,
        show: show,
      },
    }).then(function(modal) {
      modal.element.modal();
      modal.close.then(function(result) {
      });
    }).catch(function(error) {
      console.log(error);
    });
  };
})
.controller('EpisodeSearchCtrl', function($scope, $element, ep, show, EpisodesModel, close) {
  var episodeSearchCtrl = this;
  episodeSearchCtrl.searching = "searching";
  episodeSearchCtrl.ep = ep;
  episodeSearchCtrl.show = show;

  EpisodesModel.searchEpisode(show, ep).then(function(result){
    episodeSearchCtrl.searchResults = result;
  });

  episodeSearchCtrl.downloadResult = function(res) {
    console.log("Sending for download");
    console.log(res);
    EpisodesModel.downloadEpisode(show, ep, res)
    close();
  }

  episodeSearchCtrl.close = function(result) {
    close(result, 500);
  }
});
